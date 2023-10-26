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

#pragma once

#define VK_NO_PROTOTYPES
#include <vulkan/vulkan.h>

#ifdef __cplusplus
extern "C" {
#endif

#define VK_PROC_ADDR(FN) PFN_##FN FN = (PFN_##FN)r->procAddr(r->instance, #FN)

#define VK_PROC_ADDR_ERROR(FN)               \
	VK_PROC_ADDR(FN);                        \
	if (!FN) {                               \
		return VK_ERROR_INCOMPATIBLE_DRIVER; \
	}

typedef struct {
	VkInstance instance;
	VkSurfaceKHR surface;

	VkPhysicalDevice physicalDevice;
	VkDevice device;

	VkExtent2D surfaceExtent;
	VkSurfaceFormatKHR surfaceFormat;

	uint32_t graphicsQueueFamilyIndex;
	VkQueue graphicsQueue;

	uint32_t presentQueueFamilyIndex;
	VkQueue presentQueue;

	VkSwapchainKHR swapChain;
	VkRenderPass renderPass;
	VkCommandPool commandPool;
	VkPipelineLayout pipelineLayout;
	VkPipeline pipeline;

	uint32_t swapChainSz;
	uint32_t currentFrame;
	VkImageView* swapChainImageViews;
	VkFramebuffer* framebuffers;
	VkCommandBuffer* commandbuffers;
	VkSemaphore* imageAvailableSemaphores;
	VkSemaphore* renderFinishedSemaphores;
	VkFence* inFlightFences;

	PFN_vkGetInstanceProcAddr procAddr;
	VkDebugUtilsMessengerEXT messenger;
} renderer;

extern VkBool32 goVkLog(VkDebugUtilsMessageSeverityFlagBitsEXT,
						VkDebugUtilsMessageTypeFlagsEXT,
						VkDebugUtilsMessengerCallbackDataEXT*,
						void*);

extern VkResult vkShaderLoad(renderer*, const char*, VkShaderModule*);

extern VkResult vkInitLog(renderer*);
extern VkResult vkInitDevice(renderer*);
extern VkResult vkInitSwapChain(renderer*);
extern VkResult vkDeviceWaitIdle(renderer*);

extern void vkDraw(renderer*);

extern void vkDestroyLog(renderer* r);
extern void vkDestroyDevice(renderer* r);
extern void vkDestroySwapChain(renderer* r);

#ifdef __cplusplus
}
#endif
