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

import (
	"os"
	"time"

	"goarrg.com"
	"goarrg.com/debug"
)

type program struct {
	timer *time.Timer
}

func (p *program) Init(goarrg.PlatformInterface) error {
	p.timer = time.NewTimer(time.Millisecond * 500)

	return nil
}

func (p *program) Update(deltaTime float64) {
	select {
	case <-p.timer.C:
		err := PlaySound("test2.wav")
		if err != nil {
			debug.EPrint(err)
			os.Exit(1)
		}

		p.timer.Reset(time.Millisecond * 500)
	default:
	}
}

func (p *program) Shutdown() bool {
	return true
}

func (p *program) Destroy() {
}
