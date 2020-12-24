//+build !debug
//+build !disable_gl

/*
Copyright 2020 The goARRG Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package texture

import (
	_ "image/png"
	"reflect"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"goarrg.com/examples/shared/gl2d"
)

func TestTextureClose(t *testing.T) {
	const numSprites = 10
	sprites := make([]gl2d.Sprite, 0, numSprites)

	for i := 0; i < numSprites; i++ {
		s, err := gl2d.SpriteLoad("test.png")

		if err != nil {
			t.Fatal(err)
		}

		sprites = append(sprites, s)
	}

	refs := (*int64)(unsafe.Pointer(reflect.ValueOf(sprites[0]).FieldByName("texture").Elem().FieldByName("refs").Pointer()))

	if r := atomic.LoadInt64(refs); r != numSprites {
		t.Fatalf("Texture refs %d != %d", r, numSprites)
	}

	runtime.KeepAlive(sprites)
	//nolint
	sprites = nil
	runtime.GC()

	for t := time.Now(); time.Since(t) < time.Second; {
		if r := atomic.LoadInt64(refs); r == 1 {
			return
		}
	}

	t.Fatalf("Took too long")
}
