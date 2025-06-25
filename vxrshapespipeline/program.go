//go:build !goarrg_disable_vk
// +build !goarrg_disable_vk

/*
Copyright 2025 The goARRG Authors.

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

import "goarrg.com"

type program struct{}

func (p *program) Init(goarrg.PlatformInterface) error {
	return nil
}

func (p *program) Update(deltaTime float64) {
}

func (p *program) Shutdown() bool {
	return true
}

func (p *program) Destroy() {
}
