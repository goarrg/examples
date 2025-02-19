//go:build !goarrg_disable_vk
// +build !goarrg_disable_vk

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
	#cgo pkg-config: vulkan-headers

	#include "renderer.h"
*/
import "C"

import (
	"unsafe"

	"goarrg.com/debug"
)

//export goVkLog
func goVkLog(cMessageSeverity C.VkDebugUtilsMessageSeverityFlagBitsEXT,
	cMessageType C.VkDebugUtilsMessageTypeFlagsEXT,
	cCallbackData *C.VkDebugUtilsMessengerCallbackDataEXT,
	cUserData unsafe.Pointer,
) C.VkBool32 {
	format := "%s"
	args := []interface{}{
		C.GoString(cCallbackData.pMessage),
	}

	if cCallbackData.pMessageIdName != nil {
		format = "[%s] " + format
		args = append([]interface{}{C.GoString(cCallbackData.pMessageIdName)}, args...)
	}

	switch cMessageType {
	case C.VK_DEBUG_UTILS_MESSAGE_TYPE_GENERAL_BIT_EXT:
		format = "[VkGen] " + format
	case C.VK_DEBUG_UTILS_MESSAGE_TYPE_VALIDATION_BIT_EXT:
		format = "[VkVal] " + format
	case C.VK_DEBUG_UTILS_MESSAGE_TYPE_PERFORMANCE_BIT_EXT:
		format = "[VkPer] " + format
	}

	switch cMessageSeverity {
	case C.VK_DEBUG_UTILS_MESSAGE_SEVERITY_VERBOSE_BIT_EXT:
		debug.VPrintf(format, args...)
	case C.VK_DEBUG_UTILS_MESSAGE_SEVERITY_INFO_BIT_EXT:
		debug.IPrintf(format, args...)
	case C.VK_DEBUG_UTILS_MESSAGE_SEVERITY_WARNING_BIT_EXT:
		debug.WPrintf(format, args...)
	case C.VK_DEBUG_UTILS_MESSAGE_SEVERITY_ERROR_BIT_EXT:
		debug.EPrintf(format, args...)
	}

	return C.VK_FALSE
}
