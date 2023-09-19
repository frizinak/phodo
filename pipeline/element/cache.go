package element

import (
	"bytes"
	"fmt"
	"sort"
	"time"

	"github.com/frizinak/phodo/img48"
	"github.com/frizinak/phodo/pipeline"
	"github.com/frizinak/phodo/pipeline/element/core"
)

const CacheStorageName = "stdlib.cache"

func cacheContainer(ctx pipeline.Context) *CacheContainer {
	return ctx.Get(CacheStorageName).(*CacheContainer)
}

func Once(els ...pipeline.Element) pipeline.Element {
	return cache{
		container: NewCacheContainer(1024 * 1024 * 1024),
		hash:      dummyHash{make([]byte, 1)},
		p:         pipeline.New(els...),
	}
}

type CacheContainer struct {
	max  uint64
	size uint64
	l    map[string]*cacheEntry
}

func NewCacheContainer(max uint64) *CacheContainer {
	return &CacheContainer{max: max, l: make(map[string]*cacheEntry)}
}

func (c *CacheContainer) Get(sum []byte) (*img48.Img, bool) {
	var img *img48.Img
	v, ok := c.l[string(sum)]
	if ok {
		ok = len(v.sum) != 0 && bytes.Equal(v.sum, sum)
		img = v.Img
	}

	if ok {
		v.access = time.Now()
	}

	return img, ok
}

func (c *CacheContainer) Set(sum []byte, img *img48.Img) {
	e := &cacheEntry{
		access: time.Now(),
		sum:    sum,
		Img:    img,
	}

	k := string(sum)

	if e, ok := c.l[k]; ok {
		c.size -= e.Size()
	}

	size := e.Size()
	if size > c.max {
		return
	}

	c.size += size
	c.l[k] = e

	c.Cleanup()
}

func (c *CacheContainer) Cleanup() {
	if c.size <= c.max {
		return
	}

	list := make([]*cacheEntry, 0, len(c.l))
	for _, i := range c.l {
		list = append(list, i)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[j].access.Before(list[i].access)
	})

	m := make(map[string]*cacheEntry, len(c.l))
	var size uint64
	ix := 0
	for size < c.max {
		e := list[ix]
		m[string(e.sum)] = e
		size += e.Size()
		ix++
	}

	c.size = size
	c.l = m
}

func (c *CacheContainer) Clear() {
	c.l = make(map[string]*cacheEntry)
	c.size = 0
}

type dummyHash struct {
	v []byte
}

func (d dummyHash) Value() []byte { return d.v }

type cacheEntry struct {
	access time.Time
	sum    []byte
	*img48.Img
}

func (c cacheEntry) Size() uint64 { return uint64(len(c.Img.Pix)) * 2 }

type cache struct {
	container *CacheContainer

	hash pipeline.Hash
	p    *pipeline.Pipeline
}

func (c cache) Name() string { return "cache" }

func (c cache) Help() [][2]string {
	return [][2]string{
		{
			fmt.Sprintf("%s([element1] [element2] ...[elementN])", c.Name()),
			"Create a pipeline out of its arguments and caches the result.",
		},
	}
}

func (c cache) Encode(w pipeline.Writer) error {
	return c.p.Encode(w)
}

func (c cache) Decode(r pipeline.Reader) (interface{}, error) {
	p, err := (&pipeline.Pipeline{}).Decode(r)
	if err != nil {
		return c, err
	}

	c.p = p.(*pipeline.Pipeline)
	c.hash = r.Hash()

	return c, err
}

func (c cache) Do(ctx pipeline.Context, img *img48.Img) (*img48.Img, error) {
	ctx.Mark(c)

	if c.container == nil {
		c.container = cacheContainer(ctx)
	}

	hash := c.hash.Value()
	if img, ok := c.container.Get(hash); ok {
		return core.ImageCopyDiscard(img), nil
	}

	img, err := c.p.Do(ctx, img)
	if err == nil {
		c.container.Set(hash, core.ImageCopyDiscard(img))
	}

	return img, err
}
