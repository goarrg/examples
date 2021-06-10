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

package main

import (
	"os"

	"goarrg.com/debug"
	"goarrg.com/examples/shared/gl2d"
	"goarrg.com/input"
)

type program struct {
	sprite gl2d.Sprite
}

// Init is called only once, engine will exit if error
func (p *program) Init() error {
	// loads a sprite
	s, err := gl2d.SpriteLoad("test.png")

	// if error exit
	if debug.LogErr(err) {
		os.Exit(1)
	}

	p.sprite = s
	return nil
}

/*
	Update is called every frame, deltaTime is time between frames as reported
	by the renderer.
*/
func (p *program) Update(deltaTime float64) {
	// gets the mouse device
	mouse := input.DeviceOfType(input.DeviceTypeMouse)

	// gets state for mouse pos and assert it to a input.Coords
	mousePos := mouse.StateFor(input.MouseMotion).(input.Coords)

	// sets sprite pos to mouse pos
	p.sprite.SetPos(gl2d.ScreenPosToWorld(mousePos.Point3f64))
	// draw sprite
	gl2d.Render(p.sprite)
}

/*
	Shutdown is called when goarrg.Shutdown() was signaled and after the
	main loop has finished. Returning false will cancel the shutdown unless
	a SIGINT was received then Shutdown() will not be called and will not be able
	to avoid termination.
*/
func (p *program) Shutdown() bool {
	return true
}

// Destroy is called when it is time to terminate
func (p *program) Destroy() {
}
