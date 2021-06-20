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

package gl2d

/*
	#cgo linux LDFLAGS: -lGL -lGLU
	#cgo windows LDFLAGS: -lopengl32 -lglu32
	#include <GL/gl.h>
	#include <GL/glu.h>
*/
import "C"
import (
	"image"
	"runtime"
	"sync/atomic"
	"unsafe"

	"goarrg.com/asset"
	"goarrg.com/debug"
	"goarrg.com/gmath"
)

type texture struct {
	refs       *int64
	id         C.GLuint
	filename   string
	resolution gmath.Vector3i
}

func textureLoad(file string) (*texture, error) {
	Renderer.textureLock.RLock()

	if t, ok := Renderer.textures[file]; ok {
		atomic.AddInt64(t.refs, 1)
		Renderer.textureLock.RUnlock()
		runtime.SetFinalizer(&t, (*texture).close)
		return &t, nil
	}

	Renderer.textureLock.RUnlock()
	Renderer.textureLock.Lock()
	defer Renderer.textureLock.Unlock()

	// be 100% sure it wasn't added between the RUnlock() and Lock()
	if t, ok := Renderer.textures[file]; ok {
		atomic.AddInt64(t.refs, 1)
		runtime.SetFinalizer(&t, (*texture).close)
		return &t, nil
	}

	a, err := asset.Load(file)

	if err != nil {
		return nil, debug.ErrorWrapf(err, "Failed to load texture")
	}

	img, _, err := image.Decode(a.Reader())

	if err != nil {
		return nil, debug.ErrorWrapf(err, "Failed to load texture")
	}

	switch t := img.(type) {
	case *image.RGBA:
	case *image.NRGBA:
	default:
		return nil, debug.Errorf("Unsupported image format %T", t)
	}

	t := texture{
		refs:     new(int64),
		filename: a.Filename(),
		resolution: gmath.Vector3i{
			X: img.Bounds().Dx(),
			Y: img.Bounds().Dy(),
		},
	}

	Renderer.runAsync(func() {
		C.glGenTextures(1, &t.id)
		C.glBindTexture(C.GL_TEXTURE_2D, t.id)

		C.glTexParameterf(C.GL_TEXTURE_2D, C.GL_TEXTURE_MAG_FILTER, C.GL_LINEAR)
		C.glTexParameterf(C.GL_TEXTURE_2D, C.GL_TEXTURE_MIN_FILTER, C.GL_LINEAR)

		// C.glPixelStorei(C.GL_UNPACK_ALIGNMENT, rowAlign) // 1, 2, 4, 8

		switch img := img.(type) {
		case *image.RGBA:
			C.glTexImage2D(C.GL_TEXTURE_2D, 0, C.GL_RGBA8,
				C.int(img.Bounds().Dx()), C.int(img.Bounds().Dy()), 0, C.GL_RGBA,
				C.GL_UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))
		case *image.NRGBA:
			C.glTexImage2D(C.GL_TEXTURE_2D, 0, C.GL_RGBA8,
				C.int(img.Bounds().Dx()), C.int(img.Bounds().Dy()), 0, C.GL_RGBA,
				C.GL_UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))
		default:
			C.glDeleteTextures(1, &t.id)
		}

		glErr := C.glGetError()

		if glErr != C.GL_NO_ERROR {
			for ; glErr != C.GL_NO_ERROR; glErr = C.glGetError() {
				debug.LogE("Error during processing texture %s %v", t.filename, debug.Errorf(C.GoString((*C.char)(unsafe.Pointer(C.gluErrorString(glErr))))))
			}
			return
		}

		Renderer.textureLock.Lock()
		if _, ok := Renderer.textures[t.filename]; ok {
			Renderer.textures[file] = t
		}
		Renderer.textureLock.Unlock()
	})

	(*t.refs) = 1
	Renderer.textures[file] = t

	runtime.SetFinalizer(&t, (*texture).close)
	return &t, err
}

func (t *texture) close() {
	debug.LogI("Decrementing reference for texture %q", t.filename)
	if atomic.AddInt64(t.refs, -1) <= 0 {
		Renderer.textureLock.Lock()
		defer Renderer.textureLock.Unlock()

		if t, ok := Renderer.textures[t.filename]; ok {
			debug.LogI("Deleting unused texture %q", t.filename)
			Renderer.runAsync(func() {
				C.glDeleteTextures(1, &t.id)
			})
			delete(Renderer.textures, t.filename)
		}
	}
}
