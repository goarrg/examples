//go:build !goarrg_disable_gl && !goarrg_disable_vk
// +build !goarrg_disable_gl,!goarrg_disable_vk

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
	"goarrg.com"
	"goarrg.com/debug"
)

type renderer struct {
	gl2d goarrg.GLRenderer
}

var (
	_ goarrg.VkRenderer = &renderer{}
	_ goarrg.GLRenderer = &renderer{}
)

func (r *renderer) VkConfig() goarrg.VkConfig {
	return goarrg.VkConfig{
		API:        (((1) << 22) | ((1) << 12) | (0)),
		Layers:     []string{},
		Extensions: []string{"VK_EXT_debug_utils"},
	}
}

func (r *renderer) GLConfig() goarrg.GLConfig {
	return r.gl2d.GLConfig()
}

func (r *renderer) VkInit(_ goarrg.PlatformInterface, vkInstance goarrg.VkInstance) error {
	return debug.Errorf("Test GL fallback")
}

func (r *renderer) GLInit(p goarrg.PlatformInterface, vkInstance goarrg.GLInstance) error {
	return r.gl2d.GLInit(p, vkInstance)
}

func (r *renderer) Update() {
}

func (r *renderer) Draw() float64 {
	return r.gl2d.Draw()
}

func (r *renderer) Resize(x int, y int) {
	r.gl2d.Resize(x, y)
}

func (r *renderer) Destroy() {
	r.gl2d.Destroy()
}
