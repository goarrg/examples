//go:build !goarrg_disable_gl
// +build !goarrg_disable_gl

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

/*
	#cgo pkg-config: --static sdl3
*/
import "C"

import (
	_ "image/png"
	"os"

	"goarrg.com"
	"goarrg.com/debug"
	"goarrg.com/examples/shared/gl2d"
	"goarrg.com/gmath"
	"goarrg.com/platform/sdl"
)

func main() {
	// uncomment to change log level
	// debug.LogSetLevel(debug.LogLevelError)

	// configure sdl
	err := sdl.Setup(sdl.Config{
		// configure window size and position
		Window: sdl.WindowConfig{
			Rect: gmath.Rectint{X: -1, Y: -1, W: 800, H: 600},
		},
	})
	// if error, exit
	if err != nil {
		debug.EPrint(err)
		os.Exit(1)
	}

	// runs goarrg with given config
	err = goarrg.Run(goarrg.Config{
		Platform: sdl.Platform,  // platform selection
		Audio:    nil,           // audio mixer selection
		Renderer: gl2d.Renderer, // renderer selection
		Program:  &program{},    // program (game logic) selection
	})
	// if error, exit
	if err != nil {
		debug.EPrint(err)
		os.Exit(1)
	}
}
