/*
Copyright 2026 The goARRG Authors.

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

#include <stddef.h>
#include <stdint.h>

#include <vkm/vkm.h>
#include <vkm/std/utility.hpp>
#include <vkm/std/vector.hpp>

#include <SDL3/SDL.h>
#include <SDL3/SDL_main.h>
#include <SDL3/SDL_vulkan.h>

#define NUM_FIF 2

void logger(vkm_logLevel l,
			size_t numTags,
			const vkm_string* tags,
			vkm_string s) {
	vkm::std::stringbuilder builder;
	for (size_t i = 0; i < numTags; i++) {
		builder << "[" << tags[i] << "] ";
	}
	builder << s;
	switch (l) {
		case VKM_LOG_LEVEL_VERBOSE:
			SDL_Log("[V] %s", builder.cStr());
			break;
		case VKM_LOG_LEVEL_INFO:
			SDL_Log("[I] %s", builder.cStr());
			break;
		case VKM_LOG_LEVEL_WARN:
			SDL_Log("[W] %s", builder.cStr());
			break;
		case VKM_LOG_LEVEL_ERROR:
			SDL_Log("[E] %s", builder.cStr());
			break;
		default:
			SDL_Log("Unknown log level: %d\n%s", l, s.ptr);
			break;
	}
}

int main(int argc, char** argv) {
	if (!SDL_Init(SDL_INIT_VIDEO)) {
		SDL_Log("SDL could not initialize! SDL error: %s\n", SDL_GetError());
		return 1;
	}
	DEFER([&]() { SDL_Quit(); });
	auto* window = SDL_CreateWindow("VKM", 800, 600, SDL_WINDOW_VULKAN);
	if (window == nullptr) {
		SDL_Log("Window could not be created! SDL error: %s\n", SDL_GetError());
		return 1;
	}
	DEFER([&]() { SDL_DestroyWindow(window); });

	VkInstance instance;
	VkSurfaceKHR surface = VK_NULL_HANDLE;
	vkm_device device = nullptr;
	vkm_context context = nullptr;
	{
		VkResult ret = vkm_init(vkm_initInfo{
			.loggerFn = logger,
			.procAddr = reinterpret_cast<PFN_vkGetInstanceProcAddr>(
				SDL_Vulkan_GetVkGetInstanceProcAddr()),
		});
		if (ret != VK_INCOMPLETE) {
			SDL_Log("Failed to init vkm: %d", ret);
			return 1;
		}

		vkm_initializer initializer;
		vkm_createInitializer(
			vkm_initializerCreateInfo{
				.api = VKM_VK_API,
				.preferType = VKM_INITIALIZER_PREFER_INTEGRATED,
			},
			&initializer);
		DEFER([&]() { vkm_destroyInitializer(initializer); });
		{
			uint32_t numExt;
			auto exts = SDL_Vulkan_GetInstanceExtensions(&numExt);
			for (size_t i = 0; i < numExt; i++) {
				vkm_initializer_findExtension(initializer, VK_TRUE,
											  VKM_MAKE_STRING(exts[i]));
			}
		}
		auto features13 = VkPhysicalDeviceVulkan13Features{
			.sType = VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VULKAN_1_3_FEATURES,
			.dynamicRendering = VK_TRUE,
		};
		vkm_initializer_findFeature(initializer, VK_TRUE, &features13);
		auto swapchainMaint1 = VkPhysicalDeviceSwapchainMaintenance1FeaturesEXT{
			.sType =
				VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_SWAPCHAIN_MAINTENANCE_1_FEATURES_EXT,
			.swapchainMaintenance1 = VK_TRUE,
		};
		vkm_initializer_findExtension(
			initializer, VK_FALSE,
			VKM_MAKE_STRING(VK_EXT_SWAPCHAIN_MAINTENANCE_1_EXTENSION_NAME));
		vkm_initializer_findFeature(initializer, VK_FALSE, &swapchainMaint1);
		vkm_initializer_findGraphicsQueue(initializer,
										  vkm_initializer_queueCreateInfo{
											  .min = 1,
										  });

		ret = vkm_initializer_createInstance(initializer, &instance);
		if (ret != VK_SUCCESS) {
			SDL_Log("Failed to create instance: %d", ret);
			return 1;
		}
		if (!SDL_Vulkan_CreateSurface(window, instance, nullptr, &surface)) {
			SDL_Log("Failed to create surface: %s", SDL_GetError());
			return 1;
		}
		vkm_initializer_findPresentationSupport(initializer, surface);
		ret = vkm_initializer_createDevice(initializer, &device);
		if (ret != VK_SUCCESS) {
			SDL_Log("Failed to create device: %d", ret);
			return 1;
		}

		vkm_initializer_queueInfo qInfo;
		vkm_initializer_getGraphicsQueueInfo(initializer, &qInfo);
		if (qInfo.count == 0) {
			SDL_Log("Failed to retrieve graphics queue");
			return 1;
		}
		vkm_createContext(device, VKM_MAKE_STRING("main_queue"),
						  vkm_contextCreateInfo{
							  .queueFamily = qInfo.family,
							  .maxPendingFrames = NUM_FIF,
						  },
						  &context);
	}
	DEFER([&]() { vkm_shutdown(); });
	DEFER([&]() { SDL_Vulkan_DestroySurface(instance, surface, nullptr); });
	DEFER([&]() { vkm_destroyDevice(device); });
	DEFER([&]() { vkm_destroyContext(context); });

	vkm_swapchain swapchain = nullptr;
	{
		int w, h;
		if (!SDL_GetWindowSizeInPixels(window, &w, &h)) {
			SDL_Log("Failed to get surface size: %s", SDL_GetError());
			return 1;
		}
		VkResult ret = vkm_createSwapchain(
			device, VKM_MAKE_STRING("main_window"),
			vkm_swapchainCreateInfo{
				.targetSurface = surface,
				.extent = VkExtent2D{static_cast<uint32_t>(w),
									 static_cast<uint32_t>(h)},
				.requiredUsage = VK_IMAGE_USAGE_TRANSFER_DST_BIT,
				.preferredImageCount = NUM_FIF + 1,
			},
			&swapchain);
		if (ret != VK_SUCCESS) {
			SDL_Log("Failed to init swapchain: %d", ret);
			return 1;
		}
	}
	DEFER([&]() { vkm_destroySwapchain(swapchain); });

	bool quit = false;
	while (!quit) {
		SDL_Event e;
		while (SDL_PollEvent(&e) == true) {
			if (e.type == SDL_EVENT_QUIT) {
				quit = true;
			}
		}

		vkm_context_begin(context, VKM_MAKE_STRING("main"));
		VkCommandBuffer cb;
		vkm_context_beginCommandBuffer(context, VKM_MAKE_STRING("main"),
									   vkm_context_commandBufferBeginInfo{},
									   &cb);

		VkResult ret;
		vkm_swapcain_image image;
		{
			vkm_swapchain_acquireInfo acquireInfo{
				.swapchain = swapchain,
				.pImage = &image,
				.pResult = &ret,
			};
			vkm_context_acquireSwapchain(context, 1, &acquireInfo);
			if (ret != VK_SUCCESS) {
				SDL_Log("Failed to acquire swapchain: %d", ret);
				return 1;
			}
		}

		VkImageSubresourceRange range = {
			.aspectMask = VK_IMAGE_ASPECT_COLOR_BIT,
			.levelCount = 1,
			.layerCount = 1,
		};

		{
			VkImageMemoryBarrier2 barrier = {
				.sType = VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER_2,
				.dstStageMask = VK_PIPELINE_STAGE_2_ALL_TRANSFER_BIT,
				.dstAccessMask = VK_ACCESS_2_MEMORY_WRITE_BIT,
				.oldLayout = VK_IMAGE_LAYOUT_UNDEFINED,
				.newLayout = VK_IMAGE_LAYOUT_GENERAL,
				.image = image.vkImage,
				.subresourceRange = range,
			};
			VkDependencyInfo dInfo = {
				.sType = VK_STRUCTURE_TYPE_DEPENDENCY_INFO,
				.imageMemoryBarrierCount = 1,
				.pImageMemoryBarriers = &barrier,
			};
			VKM_DEVICE_VKFN(device, vkCmdPipelineBarrier2)(cb, &dInfo);
		}

		VkClearColorValue color = {.float32{0.5, 0.5, 0.5, 1}};

		VKM_DEVICE_VKFN(device, vkCmdClearColorImage)(
			cb, image.vkImage, VK_IMAGE_LAYOUT_GENERAL, &color, 1, &range);

		{
			VkImageMemoryBarrier2 barrier = {
				.sType = VK_STRUCTURE_TYPE_IMAGE_MEMORY_BARRIER_2,
				.srcStageMask = VK_PIPELINE_STAGE_2_ALL_TRANSFER_BIT,
				.srcAccessMask = VK_ACCESS_2_MEMORY_WRITE_BIT,
				.oldLayout = VK_IMAGE_LAYOUT_GENERAL,
				.newLayout = VK_IMAGE_LAYOUT_PRESENT_SRC_KHR,
				.image = image.vkImage,
				.subresourceRange = range,
			};
			VkDependencyInfo dInfo = {
				.sType = VK_STRUCTURE_TYPE_DEPENDENCY_INFO,
				.imageMemoryBarrierCount = 1,
				.pImageMemoryBarriers = &barrier,
			};
			VKM_DEVICE_VKFN(device, vkCmdPipelineBarrier2)(cb, &dInfo);
		}

		vkm_swapchain_presentInfo presentInfo{
			.swapchain = swapchain,
			.pResult = &ret,
		};
		vkm_context_endCommandBuffer(context, vkm_context_commandBufferEndInfo{
												  .numPrsentInfos = 1,
												  .pPresentInfos = &presentInfo,
											  });
		vkm_context_end(context);

		if (ret != VK_SUCCESS) {
			SDL_Log("Failed to submit frame: %d", ret);
			return 1;
		}
	}

	return 0;
}
