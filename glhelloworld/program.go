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
	"time"

	"goarrg.com"
	"goarrg.com/debug"
	"goarrg.com/examples/shared/gl2d"
	"goarrg.com/input"
)

type program struct {
	sprite gl2d.Sprite
}

// Init function is called only once,
// engine will exit if error
func (p *program) Init() error {
	// auto shutdown after 5 seconds
	time.AfterFunc(time.Second*5, func() {
		goarrg.Shutdown()
	})

	// loads a sprite
	s, err := gl2d.SpriteLoad("test.png")

	// if error exit
	if debug.LogErr(err) {
		os.Exit(1)
	}

	p.sprite = s
	return nil
}

// Update function is called every frame
// deltaTime is time between frames as reported by the renderer
// driver is snapshot of input events for current frame
func (p *program) Update(deltaTime float64, driver input.Snapshot) {
	// gets the mouse device
	mouse, err := driver.Device(input.DeviceTypeMouse)

	if debug.LogErr(err) {
		os.Exit(1)
	}

	// gets state for mouse pos and assert it to a input.Coords
	mousePos := mouse.StateFor(input.MouseMotion).(input.Coords)

	// sets sprite pos to mouse pos
	p.sprite.SetPos(mousePos.Point3f64)
	// draw sprite
	gl2d.Render(p.sprite)
}

// Shutdown function is called when goarrg.Shutdown() is called
// if function returns false, shutdown is avoided
// this is useful to have user confirmation or to make sure everything's flushed to disk
// you have to call goarrg.Shutdown() if you returned false and want to shutdown
func (p *program) Shutdown() bool {
	return true
}

// Destroy is called after shutdown confirmation and all shutdown functions
func (p *program) Destroy() {
}
