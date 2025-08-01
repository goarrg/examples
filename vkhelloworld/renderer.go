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

/*
	#cgo pkg-config: --static vulkan-headers sdl3
	#cgo LDFLAGS: -lstdc++

	#include "renderer.h"

	static inline void vkInitPlatform(renderer * r, uintptr_t instance,
									  uint64_t surface, uintptr_t procAddr) {
		r->instance = (VkInstance)instance;
		r->surface = (VkSurfaceKHR)surface;
		r->procAddr = (PFN_vkGetInstanceProcAddr)procAddr;
	}
*/
import "C"

import (
	"goarrg.com"
	"goarrg.com/debug"
)

type renderer struct {
	cRenderer C.renderer
}

func (r *renderer) VkConfig() goarrg.VkConfig {
	return goarrg.VkConfig{
		API:        ((1) << 22) | ((1) << 12),
		Layers:     []string{},
		Extensions: []string{"VK_EXT_debug_utils"},
	}
}

func (r *renderer) VkInit(_ goarrg.PlatformInterface, vkInstance goarrg.VkInstance) error {
	s, err := vkInstance.CreateSurface()
	if err != nil {
		return debug.ErrorWrapf(err, "Failed to create surface")
	}
	C.vkInitPlatform(&r.cRenderer,
		C.uintptr_t(vkInstance.Uintptr()),
		C.uint64_t(s),
		C.uintptr_t(vkInstance.ProcAddr()),
	)

	if cErr := C.vkInitLog(&r.cRenderer); cErr != C.VK_SUCCESS {
		return debug.ErrorWrapf(debug.Errorf(vkResultStr(cErr)), "Failed to init vk logger")
	}

	if cErr := C.vkInitDevice(&r.cRenderer); cErr != C.VK_SUCCESS {
		return debug.ErrorWrapf(debug.Errorf(vkResultStr(cErr)), "Failed to init vk device")
	}

	if cErr := C.vkInitSwapChain(&r.cRenderer); cErr != C.VK_SUCCESS {
		return debug.ErrorWrapf(debug.Errorf(vkResultStr(cErr)), "Failed to init vk swap chain")
	}
	return nil
}

func (r *renderer) Update() {
}

func (r *renderer) Draw() float64 {
	C.vkDraw(&r.cRenderer)
	return 0
}

func (r *renderer) Resize(int, int) {
	C.vkDeviceWaitIdle(&r.cRenderer)
	C.vkDestroySwapChain(&r.cRenderer)
	if cErr := C.vkInitSwapChain(&r.cRenderer); cErr != C.VK_SUCCESS {
		panic(debug.ErrorWrapf(debug.Errorf(vkResultStr(cErr)), "Failed to init vk swap chain"))
	}
}

func (r *renderer) Destroy() {
	C.vkDeviceWaitIdle(&r.cRenderer)

	// Destroy in reverse order of init
	C.vkDestroySwapChain(&r.cRenderer)
	C.vkDestroyDevice(&r.cRenderer)
	C.vkDestroyLog(&r.cRenderer)
}
