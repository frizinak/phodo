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
		container: NewCacheContainer(10),
		hash:      dummyHash{make([]byte, 1)},
		els:       els,
	}
}

type CacheContainer struct {
	max int
	l   map[string]*cacheEntry
}

func NewCacheContainer(max int) *CacheContainer {
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
	c.l[string(sum)] = &cacheEntry{
		access: time.Now(),
		sum:    sum,
		Img:    img,
	}

	c.Cleanup()
}

func (c *CacheContainer) Cleanup() {
	if len(c.l) <= c.max {
		return
	}
	list := make([]*cacheEntry, 0, len(c.l))
	for _, i := range c.l {
		list = append(list, i)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].access.Before(list[j].access)
	})

	m := make(map[string]*cacheEntry, c.max)
	for i := 0; i < c.max; i++ {
		m[string(list[i].sum)] = list[i]
	}
	c.l = m
}

func (c *CacheContainer) Clear() {
	c.l = make(map[string]*cacheEntry)
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

type cache struct {
	container *CacheContainer

	hash pipeline.Hash
	els  []pipeline.Element
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
	for _, el := range c.els {
		if err := w.Element(el); err != nil {
			return err
		}
	}
	return nil
}

func (c cache) Decode(r pipeline.Reader) (pipeline.Element, error) {
	ne := r.Len()
	c.els = make([]pipeline.Element, ne)
	for i := range c.els {
		var err error
		c.els[i], err = r.Element()
		if err != nil {
			return c, err
		}
	}

	c.hash = r.Hash()

	return c, nil
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

	img, err := pipeline.New(c.els...).Do(ctx, img)
	if err == nil {
		c.container.Set(hash, core.ImageCopyDiscard(img))
	}

	return img, err
}
