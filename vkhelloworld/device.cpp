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

#include <stdlib.h>
#include <string.h>

#include "defer.hpp"
#include "renderer.h"

static char const* const deviceExtensions[] = {"VK_KHR_swapchain"};
static uint32_t const numDeviceExtensions =
	sizeof(deviceExtensions) / sizeof(char*);

VkResult vkInitDevice(renderer* r) {
	VK_PROC_ADDR_ERROR(vkEnumeratePhysicalDevices);
	VK_PROC_ADDR_ERROR(vkEnumerateDeviceExtensionProperties);
	VK_PROC_ADDR_ERROR(vkGetPhysicalDeviceQueueFamilyProperties);
	VK_PROC_ADDR_ERROR(vkGetPhysicalDeviceSurfaceSupportKHR);
	VK_PROC_ADDR_ERROR(vkGetPhysicalDeviceSurfaceFormatsKHR);
	VK_PROC_ADDR_ERROR(vkCreateDevice);
	VK_PROC_ADDR_ERROR(vkGetDeviceQueue);
	VK_PROC_ADDR_ERROR(vkCreateCommandPool);

	uint32_t numDevices = 0;
	VkResult ret = vkEnumeratePhysicalDevices(r->instance, &numDevices, NULL);

	if (ret != VK_SUCCESS) {
		return ret;
	}

	if (numDevices <= 0) {
		return VK_ERROR_INCOMPATIBLE_DRIVER;
	}

	VkPhysicalDevice* devices =
		(VkPhysicalDevice*)calloc(numDevices, sizeof(VkPhysicalDevice));
	DEFER([&]() { free(devices); });
	ret = vkEnumeratePhysicalDevices(r->instance, &numDevices, devices);

	if (ret != VK_SUCCESS) {
		return ret;
	}

	uint32_t numExtensions = 0;
	VkExtensionProperties* extensions = NULL;
	DEFER([&]() { free(extensions); });

	uint32_t numQueueFamilies = 0;
	VkQueueFamilyProperties* queueFamilies = NULL;
	DEFER([&]() { free(queueFamilies); });

	uint32_t numSurfaceFormats = 0;
	VkSurfaceFormatKHR* surfaceFormats = NULL;
	DEFER([&]() { free(surfaceFormats); });

	for (uint32_t i = 0; i < numDevices; i++) {
		r->physicalDevice = devices[i];

		{
			uint32_t wantExtensions = numDeviceExtensions;
			uint32_t haveExtensions = 0;

			ret = vkEnumerateDeviceExtensionProperties(r->physicalDevice, NULL,
													   &haveExtensions, NULL);

			if (ret != VK_SUCCESS) {
				return ret;
			}

			if (haveExtensions <= 0) {
				continue;
			}

			if (haveExtensions > numExtensions) {
				numExtensions = haveExtensions;
				free(extensions);
				extensions = (VkExtensionProperties*)calloc(
					numExtensions, sizeof(VkExtensionProperties));
			}

			ret = vkEnumerateDeviceExtensionProperties(
				r->physicalDevice, NULL, &haveExtensions, extensions);

			if (ret != VK_SUCCESS) {
				return ret;
			}

			for (uint32_t j = 0; j < haveExtensions; j++) {
				for (uint32_t k = 0; k < numDeviceExtensions; k++) {
					if (!strcmp(extensions[j].extensionName,
								deviceExtensions[k])) {
						wantExtensions--;
						break;
					}
				}
			}

			if (wantExtensions) {
				continue;
			}
		}

		{
			uint32_t haveQueueFamilies = 0;
			vkGetPhysicalDeviceQueueFamilyProperties(r->physicalDevice,
													 &haveQueueFamilies, NULL);

			if (haveQueueFamilies > numQueueFamilies) {
				numQueueFamilies = haveQueueFamilies;
				free(queueFamilies);
				queueFamilies = (VkQueueFamilyProperties*)calloc(
					numQueueFamilies, sizeof(VkQueueFamilyProperties));
			}

			vkGetPhysicalDeviceQueueFamilyProperties(
				r->physicalDevice, &haveQueueFamilies, queueFamilies);

			r->graphicsQueueFamilyIndex = 0;
			r->presentQueueFamilyIndex = 0;

			for (uint32_t j = 0; j < haveQueueFamilies; j++) {
				if (queueFamilies[j].queueFlags & VK_QUEUE_GRAPHICS_BIT) {
					r->graphicsQueueFamilyIndex = j + 1;
				}

				VkBool32 presentSupport = VK_FALSE;
				ret = vkGetPhysicalDeviceSurfaceSupportKHR(
					r->physicalDevice, j, r->surface, &presentSupport);

				if (ret != VK_SUCCESS) {
					return ret;
				}

				if (presentSupport) {
					r->presentQueueFamilyIndex = j + 1;
				}

				if (r->graphicsQueueFamilyIndex && r->presentQueueFamilyIndex &&
					r->graphicsQueueFamilyIndex == r->presentQueueFamilyIndex) {
					break;
				}
			}

			if (!r->graphicsQueueFamilyIndex || !r->presentQueueFamilyIndex) {
				continue;
			}

			r->graphicsQueueFamilyIndex--;
			r->presentQueueFamilyIndex--;
		}

		{
			uint32_t haveSurfaceFormats;
			VkResult ret = vkGetPhysicalDeviceSurfaceFormatsKHR(
				r->physicalDevice, r->surface, &haveSurfaceFormats, NULL);

			if (ret != VK_SUCCESS) {
				return ret;
			}

			if (haveSurfaceFormats > numSurfaceFormats) {
				numSurfaceFormats = haveSurfaceFormats;
				free(surfaceFormats);
				surfaceFormats = (VkSurfaceFormatKHR*)calloc(
					numSurfaceFormats, sizeof(VkSurfaceFormatKHR));
			}

			ret = vkGetPhysicalDeviceSurfaceFormatsKHR(
				r->physicalDevice, r->surface, &haveSurfaceFormats,
				surfaceFormats);

			if (ret != VK_SUCCESS) {
				return ret;
			}

			uint32_t j = 0;
			for (; j < haveSurfaceFormats; j++) {
				if (surfaceFormats[j].format == VK_FORMAT_B8G8R8A8_SRGB &&
					surfaceFormats[j].colorSpace ==
						VK_COLOR_SPACE_SRGB_NONLINEAR_KHR) {
					r->surfaceFormat = surfaceFormats[j];
					break;
				}
			}

			if (j == haveSurfaceFormats) {
				continue;
			}
		}

		{
			uint32_t numQueueCreateInfos =
				2 - (r->graphicsQueueFamilyIndex == r->presentQueueFamilyIndex);
			VkDeviceQueueCreateInfo* queueCreateInfos =
				(VkDeviceQueueCreateInfo*)calloc(
					numQueueCreateInfos, sizeof(VkDeviceQueueCreateInfo));
			DEFER([&]() { free(queueCreateInfos); });

			float queuePriority = 1.0f;
			VkDeviceQueueCreateInfo queueCreateInfo = {};
			queueCreateInfo.sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO;
			queueCreateInfo.queueFamilyIndex = r->graphicsQueueFamilyIndex;
			queueCreateInfo.queueCount = 1;
			queueCreateInfo.pQueuePriorities = &queuePriority;
			queueCreateInfos[0] = queueCreateInfo;

			if (numQueueCreateInfos == 2) {
				VkDeviceQueueCreateInfo queueCreateInfo = {};
				queueCreateInfo.sType =
					VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO;
				queueCreateInfo.queueFamilyIndex = r->presentQueueFamilyIndex;
				queueCreateInfo.queueCount = 1;
				queueCreateInfo.pQueuePriorities = &queuePriority;
				queueCreateInfos[1] = queueCreateInfo;
			}

			VkPhysicalDeviceFeatures deviceFeatures = {};

			VkDeviceCreateInfo createInfo = {};
			createInfo.sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO;

			createInfo.queueCreateInfoCount = numQueueCreateInfos;
			createInfo.pQueueCreateInfos = queueCreateInfos;

			createInfo.pEnabledFeatures = &deviceFeatures;

			createInfo.enabledExtensionCount = numDeviceExtensions;
			createInfo.ppEnabledExtensionNames = deviceExtensions;

			ret = vkCreateDevice(r->physicalDevice, &createInfo, NULL,
								 &r->device);
			if (ret != VK_SUCCESS) {
				return ret;
			}
		}

		{
			vkGetDeviceQueue(r->device, r->graphicsQueueFamilyIndex, 0,
							 &r->graphicsQueue);
			vkGetDeviceQueue(r->device, r->presentQueueFamilyIndex, 0,
							 &r->presentQueue);
		}

		{
			VkCommandPoolCreateInfo poolInfo = {};
			poolInfo.sType = VK_STRUCTURE_TYPE_COMMAND_POOL_CREATE_INFO;
			poolInfo.queueFamilyIndex = r->graphicsQueueFamilyIndex;

			VkResult ret = vkCreateCommandPool(r->device, &poolInfo, NULL,
											   &r->commandPool);

			if (ret != VK_SUCCESS) {
				return ret;
			}
		}

		return VK_SUCCESS;
	}

	return VK_ERROR_INCOMPATIBLE_DRIVER;
}

VkResult vkDeviceWaitIdle(renderer* r) {
	VK_PROC_ADDR(vkDeviceWaitIdle);
	return vkDeviceWaitIdle(r->device);
}

void vkDestroyDevice(renderer* r) {
	VK_PROC_ADDR(vkDestroyCommandPool);
	VK_PROC_ADDR(vkDestroyDevice);

	vkDestroyCommandPool(r->device, r->commandPool, NULL);
	vkDestroyDevice(r->device, NULL);
}
