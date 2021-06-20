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
	#cgo linux LDFLAGS: -lGL
	#cgo windows LDFLAGS: -lopengl32
	#include <GL/gl.h>
*/
import "C"
import (
	"sync"
	"time"

	"goarrg.com"
	"goarrg.com/debug"
	"goarrg.com/gmath"
)

type gl2d struct {
	lock sync.Mutex
	jobs chan func()

	glInstance goarrg.GLInstance

	textureLock sync.RWMutex
	textures    map[string]texture
	sprites     []Sprite

	screenW int
	screenH int

	resW int
	resH int

	lastTime time.Time
}

var Renderer = &gl2d{
	jobs:     make(chan func(), 8),
	textures: make(map[string]texture),
}

type Config struct {
	ResW int
	ResH int
}

// setups renderer resolution
func Setup(cfg Config) error {
	if cfg.ResW <= 0 || cfg.ResH <= 0 {
		return debug.Errorf("Invalid config %+v", cfg)
	}

	Renderer.resW = cfg.ResW
	Renderer.resH = cfg.ResH

	return nil
}

// tells platform what config to use for ogl, see datatype info for options
func (r *gl2d) GLConfig() goarrg.GLConfig {
	return goarrg.GLConfig{}
}

// window and gl instance was created so now time to init the renderer
func (r *gl2d) GLInit(glInstance goarrg.GLInstance) error {
	C.glClearColor(0, 0, 0, 1)
	C.glEnable(C.GL_BLEND)
	C.glBlendFunc(C.GL_SRC_ALPHA, C.GL_ONE_MINUS_SRC_ALPHA)
	C.glDisable(C.GL_DEPTH_TEST)
	C.glEnable(C.GL_CULL_FACE)
	//C.glEnable(C.GL_MULTISAMPLE_ARB)

	if r.resW <= 0 || r.resH <= 0 {
		Renderer.resW = 800
		Renderer.resH = 600
	}

	r.glInstance = glInstance
	r.lastTime = time.Now()
	return nil
}

// time to draw, returns deltatime
func (r *gl2d) Draw() float64 {
	r.lock.Lock()
	defer r.lock.Unlock()

	t := time.Now()
	deltaTime := t.Sub(r.lastTime).Seconds()
	r.lastTime = t

lJobs:
	for start := time.Now(); time.Since(start) < time.Millisecond; {
		select {
		case j := <-r.jobs:
			j()
		default:
			break lJobs
		}
	}

	C.glClear(C.GL_COLOR_BUFFER_BIT)

	C.glViewport(0, 0, C.int(r.screenW), C.int(r.screenH))

	C.glMatrixMode(C.GL_PROJECTION)
	C.glLoadIdentity()
	C.glOrtho(0, C.double(r.resW), C.double(r.resH), 0, 0, 1)

	C.glMatrixMode(C.GL_MODELVIEW)
	C.glLoadIdentity()
	// C.glTranslatef(x, y, 0)
	// glScalef(zoom, zoom, 1)
	// glRotatef(rotation, 0, 0, 1)

	C.glEnable(C.GL_TEXTURE_2D)

	for _, s := range r.sprites {
		C.glPushMatrix()
		C.glBindTexture(C.GL_TEXTURE_2D, s.texture.id)

		// glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_REPEAT)
		// glTexParameterf(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_REPEAT)

		C.glTexParameterf(C.GL_TEXTURE_2D, C.GL_TEXTURE_WRAP_S, C.GL_CLAMP)
		C.glTexParameterf(C.GL_TEXTURE_2D, C.GL_TEXTURE_WRAP_T, C.GL_CLAMP)

		C.glTranslatef(C.float(s.Pos.X), C.float(s.Pos.Y), 0)

		C.glBegin(C.GL_TRIANGLE_STRIP)
		{
			C.glColor4f(C.float(s.Color[0]), C.float(s.Color[1]), C.float(s.Color[2]), C.float(s.Color[3]))

			// Top Left Of The Texture and Quad
			C.glTexCoord2f(C.float(s.Clip.X/s.texture.resolution.X),
				C.float(s.Clip.Y/s.texture.resolution.Y),
			)
			C.glVertex3f(0, 0, 0)

			// Bottom Left Of The Texture and Quad
			C.glTexCoord2f(C.float(s.Clip.X/s.texture.resolution.X),
				C.float((s.Clip.Y+s.Clip.H)/s.texture.resolution.Y),
			)
			C.glVertex3f(0, C.float(s.Pos.H), 0)

			// Top Right Of The Texture and Quad
			C.glTexCoord2f(C.float((s.Clip.X+s.Clip.W)/s.texture.resolution.X),
				C.float(s.Clip.Y/s.texture.resolution.Y),
			)
			C.glVertex3f(C.float(s.Pos.W), 0, 0)

			// Bottom Right Of The Texture and Quad
			C.glTexCoord2f(C.float((s.Clip.X+s.Clip.W)/s.texture.resolution.X),
				C.float((s.Clip.Y+s.Clip.H)/s.texture.resolution.Y),
			)
			C.glVertex3f(C.float(s.Pos.W), C.float(s.Pos.H), 0)
		}
		C.glEnd()

		C.glBindTexture(C.GL_TEXTURE_2D, 0)
		C.glPopMatrix()
	}

	r.sprites = r.sprites[:0]
	r.glInstance.SwapBuffers()

	return deltaTime
}

// window was resized, w and h are the drawable surface size
func (r *gl2d) Resize(w, h int) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.screenW = w
	r.screenH = h
}

// Destroy is called when it is time to terminate
func (r *gl2d) Destroy() {

}

func (r *gl2d) runAsync(f func()) {
	go func() {
		r.jobs <- f
	}()
}

func ScreenPosToWorld(pos gmath.Point3f64) gmath.Point3f64 {
	ndc := gmath.Vector3f64(pos).ScaleInverse(gmath.Vector3f64{X: float64(Renderer.screenW), Y: float64(Renderer.screenH)})
	return gmath.Point3f64(ndc.Scale(gmath.Vector3f64{X: float64(Renderer.resW), Y: float64(Renderer.resH)}))
}

// call this function to draw stuff, renderer does not draw anything you don't tell it to
func Render(sprites ...Sprite) {
	for _, s := range sprites {
		if s.texture != nil && s.Color[3] > 0 {
			Renderer.sprites = append(Renderer.sprites, s)
		}
	}
}
