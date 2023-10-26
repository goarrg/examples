//go:build !disable_vk && amd64
// +build !disable_vk,amd64

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
	"math/rand"
	"os"
	"time"

	"goarrg.com"
	"goarrg.com/debug"
	"goarrg.com/gmath"
	"goarrg.com/platform/sdl"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// debug.LogSetLevel(debug.LogLevelError)
	err := sdl.Setup(sdl.Config{
		Window: sdl.WindowConfig{
			Rect: gmath.Recti{X: -1, Y: -1, W: 800, H: 600},
		},
	})
	if err != nil {
		debug.EPrint(err)
		os.Exit(1)
	}

	err = goarrg.Run(goarrg.Config{
		Platform: sdl.Platform,
		Audio:    nil,
		Renderer: &renderer{},
		Program:  &program{},
	})

	if err != nil {
		debug.EPrint(err)
		os.Exit(1)
	}

	// Output:
}
