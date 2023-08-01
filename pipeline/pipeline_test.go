package pipeline

import (
	"testing"
)

func TestPlainNumber(t *testing.T) {
	data := map[string]bool{
		"":     false,
		"000":  false,
		"0xff": false,

		"-2.6":     true,
		".005":     true,
		"-.005":    true,
		"-2%":      true,
		"2.05%":    true,
		"200":      true,
		"1":        true,
		"200.0001": true,

		"1000*500":    false,
		"-1000-500":   false,
		"1000.50+500": false,
		"-":           false,
	}

	for k, v := range data {
		_, ok := isPlainNumber(k)
		if ok != v {
			if v {
				t.Fatalf("'%s' should have been a plain float", k)
				continue
			}
			t.Fatalf("'%s' should have not been a plain float", k)
		}
	}
}
