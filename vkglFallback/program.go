//+build !disable_gl,!disable_vk,amd64

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
)

type program struct {
	sprite gl2d.Sprite
}

func (p *program) Init() error {
	s, err := gl2d.SpriteLoad("test.png")

	if debug.LogErr(err) {
		os.Exit(1)
	}

	p.sprite = s
	return nil
}

func (p *program) Update(deltaTime float64) {
	gl2d.Render(p.sprite)
}

func (p *program) Shutdown() bool {
	return true
}

func (p *program) Destroy() {
}
