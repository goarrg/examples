//go:build !disable_gl
// +build !disable_gl

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
	"goarrg.com/debug"
	"goarrg.com/gmath"
)

type Sprite struct {
	texture *texture
	Pos     gmath.Rectf64
	Clip    gmath.Recti
	Color   [4]float32
}

// create a sprite to draw
func SpriteLoad(file string) (Sprite, error) {
	t, err := textureLoad(file)
	if err != nil {
		return Sprite{}, debug.ErrorWrapf(err, "Failed to load sprite")
	}

	return Sprite{
		texture: t,
		Pos: gmath.Rectf64{
			W: float64(t.resolution.X),
			H: float64(t.resolution.Y),
		},
		Clip: gmath.Recti{
			W: t.resolution.X,
			H: t.resolution.Y,
		},
		Color: [4]float32{
			1, 1, 1, 1,
		},
	}, nil
}

// change sprite texture
func (s *Sprite) SetTexture(file string) error {
	t, err := textureLoad(file)
	if err != nil {
		return debug.ErrorWrapf(err, "Failed to set texture")
	}

	s.texture = t

	if s.Clip == (gmath.Recti{}) {
		s.Clip = gmath.Recti{
			W: t.resolution.X,
			H: t.resolution.Y,
		}
	}

	if s.Color == ([4]float32{}) {
		s.Color = [4]float32{
			1, 1, 1, 1,
		}
	}

	return nil
}

func (s *Sprite) GetResolution() gmath.Vector3i {
	return s.texture.resolution
}

func (s *Sprite) SetPos(p gmath.Point3f64) {
	s.Pos.X = p.X
	s.Pos.Y = p.Y
}

func (s *Sprite) SetSize(v gmath.Vector3f64) {
	s.Pos.W = v.X
	s.Pos.H = v.Y
}

func (s *Sprite) Move(p gmath.Vector3f64) {
	s.Pos.X += p.X
	s.Pos.Y += p.Y
}
