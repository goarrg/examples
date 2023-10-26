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

VkResult vkInitSwapChain(renderer* r) {
	VK_PROC_ADDR_ERROR(vkGetPhysicalDeviceSurfaceCapabilitiesKHR);
	VK_PROC_ADDR_ERROR(vkGetPhysicalDeviceSurfacePresentModesKHR);
	VK_PROC_ADDR_ERROR(vkCreateSwapchainKHR);
	VK_PROC_ADDR_ERROR(vkCreateRenderPass);
	VK_PROC_ADDR_ERROR(vkCreatePipelineLayout);
	VK_PROC_ADDR_ERROR(vkCreateGraphicsPipelines);
	VK_PROC_ADDR_ERROR(vkDestroyShaderModule);
	VK_PROC_ADDR_ERROR(vkGetSwapchainImagesKHR);
	VK_PROC_ADDR_ERROR(vkAllocateCommandBuffers);
	VK_PROC_ADDR_ERROR(vkCreateImageView);
	VK_PROC_ADDR_ERROR(vkCreateFramebuffer);
	VK_PROC_ADDR_ERROR(vkBeginCommandBuffer);
	VK_PROC_ADDR_ERROR(vkCmdBeginRenderPass);
	VK_PROC_ADDR_ERROR(vkCmdBindPipeline);
	VK_PROC_ADDR_ERROR(vkCmdDraw);
	VK_PROC_ADDR_ERROR(vkCmdEndRenderPass);
	VK_PROC_ADDR_ERROR(vkEndCommandBuffer);
	VK_PROC_ADDR_ERROR(vkCreateSemaphore);
	VK_PROC_ADDR_ERROR(vkCreateFence);

	VkSurfaceCapabilitiesKHR surfaceCapabilities;
	VkPresentModeKHR presentMode;

	{
		VkResult ret = vkGetPhysicalDeviceSurfaceCapabilitiesKHR(
			r->physicalDevice, r->surface, &surfaceCapabilities);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		r->surfaceExtent = surfaceCapabilities.currentExtent;

		if (r->surfaceExtent.width == 0xFFFFFFFF ||
			r->surfaceExtent.height == 0xFFFFFFFF) {
			r->surfaceExtent = surfaceCapabilities.maxImageExtent;
		}
	}

	{
		uint32_t numPresentModes;
		VkResult ret = vkGetPhysicalDeviceSurfacePresentModesKHR(
			r->physicalDevice, r->surface, &numPresentModes, NULL);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		VkPresentModeKHR* presentModes = (VkPresentModeKHR*)calloc(
			numPresentModes, sizeof(VkPresentModeKHR));
		DEFER([&]() { free(presentModes); });

		ret = vkGetPhysicalDeviceSurfacePresentModesKHR(
			r->physicalDevice, r->surface, &numPresentModes, presentModes);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		presentMode = VK_PRESENT_MODE_FIFO_KHR;

		for (uint32_t i = 0; i < numPresentModes; i++) {
			if (presentModes[i] == VK_PRESENT_MODE_FIFO_RELAXED_KHR) {
				presentMode = presentModes[i];
				break;
			}
		}
	}

	{
		VkSwapchainCreateInfoKHR createInfo = {};
		createInfo.sType = VK_STRUCTURE_TYPE_SWAPCHAIN_CREATE_INFO_KHR;
		createInfo.surface = r->surface;

		createInfo.minImageCount = surfaceCapabilities.minImageCount;
		createInfo.imageFormat = r->surfaceFormat.format;
		createInfo.imageColorSpace = r->surfaceFormat.colorSpace;
		createInfo.imageExtent = r->surfaceExtent;
		createInfo.imageArrayLayers = 1;
		createInfo.imageUsage = VK_IMAGE_USAGE_COLOR_ATTACHMENT_BIT;
		createInfo.imageSharingMode = VK_SHARING_MODE_EXCLUSIVE;

		if (r->graphicsQueueFamilyIndex != r->presentQueueFamilyIndex) {
			uint32_t queueFamilyIndices[2] = {
				r->graphicsQueueFamilyIndex,
				r->presentQueueFamilyIndex,
			};

			createInfo.imageSharingMode = VK_SHARING_MODE_CONCURRENT;
			createInfo.queueFamilyIndexCount = 2;
			createInfo.pQueueFamilyIndices = queueFamilyIndices;
		}

		createInfo.preTransform = surfaceCapabilities.currentTransform;
		createInfo.compositeAlpha = VK_COMPOSITE_ALPHA_OPAQUE_BIT_KHR;
		createInfo.presentMode = presentMode;
		createInfo.clipped = VK_TRUE;

		createInfo.oldSwapchain = VK_NULL_HANDLE;

		VkResult ret =
			vkCreateSwapchainKHR(r->device, &createInfo, NULL, &r->swapChain);

		if (ret != VK_SUCCESS) {
			return ret;
		}
	}

	{
		VkAttachmentDescription colorAttachment = {};
		colorAttachment.format = r->surfaceFormat.format;
		colorAttachment.samples = VK_SAMPLE_COUNT_1_BIT;
		colorAttachment.loadOp = VK_ATTACHMENT_LOAD_OP_CLEAR;
		colorAttachment.storeOp = VK_ATTACHMENT_STORE_OP_STORE;
		colorAttachment.stencilLoadOp = VK_ATTACHMENT_LOAD_OP_DONT_CARE;
		colorAttachment.stencilStoreOp = VK_ATTACHMENT_STORE_OP_DONT_CARE;
		colorAttachment.initialLayout = VK_IMAGE_LAYOUT_UNDEFINED;
		colorAttachment.finalLayout = VK_IMAGE_LAYOUT_PRESENT_SRC_KHR;

		VkAttachmentReference colorAttachmentRef = {};
		colorAttachmentRef.attachment = 0;
		colorAttachmentRef.layout = VK_IMAGE_LAYOUT_COLOR_ATTACHMENT_OPTIMAL;

		VkSubpassDescription subpass = {};
		subpass.pipelineBindPoint = VK_PIPELINE_BIND_POINT_GRAPHICS;
		subpass.colorAttachmentCount = 1;
		subpass.pColorAttachments = &colorAttachmentRef;

		VkSubpassDependency dependency = {};
		dependency.srcSubpass = VK_SUBPASS_EXTERNAL;
		dependency.dstSubpass = 0;
		dependency.srcStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
		dependency.srcAccessMask = 0;
		dependency.dstStageMask = VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT;
		dependency.dstAccessMask = VK_ACCESS_COLOR_ATTACHMENT_READ_BIT |
								   VK_ACCESS_COLOR_ATTACHMENT_WRITE_BIT;

		VkRenderPassCreateInfo renderPassInfo = {};
		renderPassInfo.sType = VK_STRUCTURE_TYPE_RENDER_PASS_CREATE_INFO;
		renderPassInfo.attachmentCount = 1;
		renderPassInfo.pAttachments = &colorAttachment;
		renderPassInfo.subpassCount = 1;
		renderPassInfo.pSubpasses = &subpass;
		renderPassInfo.dependencyCount = 1;
		renderPassInfo.pDependencies = &dependency;

		VkResult ret = vkCreateRenderPass(r->device, &renderPassInfo, NULL,
										  &r->renderPass);

		if (ret != VK_SUCCESS) {
			return ret;
		}
	}

	{
		VkShaderModule vertShaderModule;
		VkShaderModule fragShaderModule;

		VkResult ret = vkShaderLoad(r, "main.vert.spv", &vertShaderModule);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		ret = vkShaderLoad(r, "main.frag.spv", &fragShaderModule);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		VkPipelineShaderStageCreateInfo vertShaderStageInfo = {};
		vertShaderStageInfo.sType =
			VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
		vertShaderStageInfo.stage = VK_SHADER_STAGE_VERTEX_BIT;
		vertShaderStageInfo.module = vertShaderModule;
		vertShaderStageInfo.pName = "main";

		VkPipelineShaderStageCreateInfo fragShaderStageInfo = {};
		fragShaderStageInfo.sType =
			VK_STRUCTURE_TYPE_PIPELINE_SHADER_STAGE_CREATE_INFO;
		fragShaderStageInfo.stage = VK_SHADER_STAGE_FRAGMENT_BIT;
		fragShaderStageInfo.module = fragShaderModule;
		fragShaderStageInfo.pName = "main";

		VkPipelineShaderStageCreateInfo shaderStages[] = {
			vertShaderStageInfo,
			fragShaderStageInfo,
		};

		VkPipelineVertexInputStateCreateInfo vertexInputInfo = {};
		vertexInputInfo.sType =
			VK_STRUCTURE_TYPE_PIPELINE_VERTEX_INPUT_STATE_CREATE_INFO;
		vertexInputInfo.vertexBindingDescriptionCount = 0;
		vertexInputInfo.vertexAttributeDescriptionCount = 0;

		VkPipelineInputAssemblyStateCreateInfo inputAssembly = {};
		inputAssembly.sType =
			VK_STRUCTURE_TYPE_PIPELINE_INPUT_ASSEMBLY_STATE_CREATE_INFO;
		inputAssembly.topology = VK_PRIMITIVE_TOPOLOGY_TRIANGLE_LIST;
		inputAssembly.primitiveRestartEnable = VK_FALSE;

		VkViewport viewport = {};
		viewport.x = 0.0f;
		viewport.y = 0.0f;
		viewport.width = (float)r->surfaceExtent.width;
		viewport.height = (float)r->surfaceExtent.height;
		viewport.minDepth = 0.0f;
		viewport.maxDepth = 1.0f;

		VkRect2D scissor = {};
		scissor.offset = (VkOffset2D){};
		scissor.extent = r->surfaceExtent;

		VkPipelineViewportStateCreateInfo viewportState = {};
		viewportState.sType =
			VK_STRUCTURE_TYPE_PIPELINE_VIEWPORT_STATE_CREATE_INFO;
		viewportState.viewportCount = 1;
		viewportState.pViewports = &viewport;
		viewportState.scissorCount = 1;
		viewportState.pScissors = &scissor;

		VkPipelineRasterizationStateCreateInfo rasterizer = {};
		rasterizer.sType =
			VK_STRUCTURE_TYPE_PIPELINE_RASTERIZATION_STATE_CREATE_INFO;
		rasterizer.depthClampEnable = VK_FALSE;
		rasterizer.rasterizerDiscardEnable = VK_FALSE;
		rasterizer.polygonMode = VK_POLYGON_MODE_FILL;
		rasterizer.lineWidth = 1.0f;
		rasterizer.cullMode = VK_CULL_MODE_BACK_BIT;
		rasterizer.frontFace = VK_FRONT_FACE_CLOCKWISE;
		rasterizer.depthBiasEnable = VK_FALSE;

		VkPipelineMultisampleStateCreateInfo multisampling = {};
		multisampling.sType =
			VK_STRUCTURE_TYPE_PIPELINE_MULTISAMPLE_STATE_CREATE_INFO;
		multisampling.sampleShadingEnable = VK_FALSE;
		multisampling.rasterizationSamples = VK_SAMPLE_COUNT_1_BIT;

		VkPipelineColorBlendAttachmentState colorBlendAttachment = {};
		colorBlendAttachment.colorWriteMask =
			VK_COLOR_COMPONENT_R_BIT | VK_COLOR_COMPONENT_G_BIT |
			VK_COLOR_COMPONENT_B_BIT | VK_COLOR_COMPONENT_A_BIT;
		colorBlendAttachment.blendEnable = VK_FALSE;

		VkPipelineColorBlendStateCreateInfo colorBlending = {};
		colorBlending.sType =
			VK_STRUCTURE_TYPE_PIPELINE_COLOR_BLEND_STATE_CREATE_INFO;
		colorBlending.logicOpEnable = VK_FALSE;
		colorBlending.logicOp = VK_LOGIC_OP_COPY;
		colorBlending.attachmentCount = 1;
		colorBlending.pAttachments = &colorBlendAttachment;
		colorBlending.blendConstants[0] = 0.0f;
		colorBlending.blendConstants[1] = 0.0f;
		colorBlending.blendConstants[2] = 0.0f;
		colorBlending.blendConstants[3] = 0.0f;

		VkPipelineLayoutCreateInfo pipelineLayoutInfo = {};
		pipelineLayoutInfo.sType =
			VK_STRUCTURE_TYPE_PIPELINE_LAYOUT_CREATE_INFO;
		pipelineLayoutInfo.setLayoutCount = 0;
		pipelineLayoutInfo.pushConstantRangeCount = 0;

		ret = vkCreatePipelineLayout(r->device, &pipelineLayoutInfo, NULL,
									 &r->pipelineLayout);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		VkGraphicsPipelineCreateInfo pipelineInfo = {};
		pipelineInfo.sType = VK_STRUCTURE_TYPE_GRAPHICS_PIPELINE_CREATE_INFO;
		pipelineInfo.stageCount = 2;
		pipelineInfo.pStages = shaderStages;
		pipelineInfo.pVertexInputState = &vertexInputInfo;
		pipelineInfo.pInputAssemblyState = &inputAssembly;
		pipelineInfo.pViewportState = &viewportState;
		pipelineInfo.pRasterizationState = &rasterizer;
		pipelineInfo.pMultisampleState = &multisampling;
		pipelineInfo.pColorBlendState = &colorBlending;
		pipelineInfo.layout = r->pipelineLayout;
		pipelineInfo.renderPass = r->renderPass;
		pipelineInfo.subpass = 0;
		pipelineInfo.basePipelineHandle = VK_NULL_HANDLE;

		ret = vkCreateGraphicsPipelines(r->device, VK_NULL_HANDLE, 1,
										&pipelineInfo, NULL, &r->pipeline);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		vkDestroyShaderModule(r->device, fragShaderModule, NULL);
		vkDestroyShaderModule(r->device, vertShaderModule, NULL);
	}

	{
		VkResult ret = vkGetSwapchainImagesKHR(r->device, r->swapChain,
											   &r->swapChainSz, NULL);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		VkImage* swapChainImages =
			(VkImage*)calloc(r->swapChainSz, sizeof(VkImage));
		DEFER([&]() { free(swapChainImages); });
		ret = vkGetSwapchainImagesKHR(r->device, r->swapChain, &r->swapChainSz,
									  swapChainImages);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		r->swapChainImageViews =
			(VkImageView*)calloc(r->swapChainSz, sizeof(VkImageView));
		r->framebuffers =
			(VkFramebuffer*)calloc(r->swapChainSz, sizeof(VkFramebuffer));
		r->commandbuffers =
			(VkCommandBuffer*)calloc(r->swapChainSz, sizeof(VkCommandBuffer));

		r->imageAvailableSemaphores =
			(VkSemaphore*)calloc(r->swapChainSz, sizeof(VkSemaphore));
		r->renderFinishedSemaphores =
			(VkSemaphore*)calloc(r->swapChainSz, sizeof(VkSemaphore));
		r->inFlightFences = (VkFence*)calloc(r->swapChainSz, sizeof(VkFence));

		VkCommandBufferAllocateInfo allocInfo = {};
		allocInfo.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_ALLOCATE_INFO;
		allocInfo.commandPool = r->commandPool;
		allocInfo.level = VK_COMMAND_BUFFER_LEVEL_PRIMARY;
		allocInfo.commandBufferCount = r->swapChainSz;

		ret =
			vkAllocateCommandBuffers(r->device, &allocInfo, r->commandbuffers);

		if (ret != VK_SUCCESS) {
			return ret;
		}

		for (uint32_t i = 0; i < r->swapChainSz; i++) {
			{
				VkImageViewCreateInfo createInfo = {};
				createInfo.sType = VK_STRUCTURE_TYPE_IMAGE_VIEW_CREATE_INFO;
				createInfo.image = swapChainImages[i];
				createInfo.viewType = VK_IMAGE_VIEW_TYPE_2D;
				createInfo.format = r->surfaceFormat.format;
				createInfo.components.r = VK_COMPONENT_SWIZZLE_IDENTITY;
				createInfo.components.g = VK_COMPONENT_SWIZZLE_IDENTITY;
				createInfo.components.b = VK_COMPONENT_SWIZZLE_IDENTITY;
				createInfo.components.a = VK_COMPONENT_SWIZZLE_IDENTITY;
				createInfo.subresourceRange.aspectMask =
					VK_IMAGE_ASPECT_COLOR_BIT;
				createInfo.subresourceRange.baseMipLevel = 0;
				createInfo.subresourceRange.levelCount = 1;
				createInfo.subresourceRange.baseArrayLayer = 0;
				createInfo.subresourceRange.layerCount = 1;

				ret = vkCreateImageView(r->device, &createInfo, NULL,
										&r->swapChainImageViews[i]);
				if (ret != VK_SUCCESS) {
					return ret;
				}
			}

			{
				VkImageView attachments[] = {r->swapChainImageViews[i]};
				VkFramebufferCreateInfo framebufferInfo = {};
				framebufferInfo.sType =
					VK_STRUCTURE_TYPE_FRAMEBUFFER_CREATE_INFO;
				framebufferInfo.renderPass = r->renderPass;
				framebufferInfo.attachmentCount = 1;
				framebufferInfo.pAttachments = attachments;
				framebufferInfo.width = r->surfaceExtent.width;
				framebufferInfo.height = r->surfaceExtent.height;
				framebufferInfo.layers = 1;

				ret = vkCreateFramebuffer(r->device, &framebufferInfo, NULL,
										  &r->framebuffers[i]);

				if (ret != VK_SUCCESS) {
					return ret;
				}

				VkCommandBufferBeginInfo beginInfo = {};
				beginInfo.sType = VK_STRUCTURE_TYPE_COMMAND_BUFFER_BEGIN_INFO;

				ret = vkBeginCommandBuffer(r->commandbuffers[i], &beginInfo);

				if (ret != VK_SUCCESS) {
					return ret;
				}
			}

			{
				VkRenderPassBeginInfo renderPassInfo = {};
				renderPassInfo.sType = VK_STRUCTURE_TYPE_RENDER_PASS_BEGIN_INFO;
				renderPassInfo.renderPass = r->renderPass;
				renderPassInfo.framebuffer = r->framebuffers[i];
				renderPassInfo.renderArea.offset = (VkOffset2D){0, 0};
				renderPassInfo.renderArea.extent = r->surfaceExtent;

				VkClearValue clearColor = {0.0f, 0.0f, 0.0f, 1.0f};
				renderPassInfo.clearValueCount = 1;
				renderPassInfo.pClearValues = &clearColor;

				vkCmdBeginRenderPass(r->commandbuffers[i], &renderPassInfo,
									 VK_SUBPASS_CONTENTS_INLINE);

				vkCmdBindPipeline(r->commandbuffers[i],
								  VK_PIPELINE_BIND_POINT_GRAPHICS, r->pipeline);

				vkCmdDraw(r->commandbuffers[i], 3, 1, 0, 0);

				vkCmdEndRenderPass(r->commandbuffers[i]);

				ret = vkEndCommandBuffer(r->commandbuffers[i]);

				if (ret != VK_SUCCESS) {
					return ret;
				}
			}

			{
				VkSemaphoreCreateInfo semaphoreInfo = {};
				semaphoreInfo.sType = VK_STRUCTURE_TYPE_SEMAPHORE_CREATE_INFO;

				VkFenceCreateInfo fenceInfo = {};
				fenceInfo.sType = VK_STRUCTURE_TYPE_FENCE_CREATE_INFO;
				fenceInfo.flags = VK_FENCE_CREATE_SIGNALED_BIT;

				ret = vkCreateSemaphore(r->device, &semaphoreInfo, NULL,
										&r->imageAvailableSemaphores[i]);

				if (ret != VK_SUCCESS) {
					return ret;
				}

				ret = vkCreateSemaphore(r->device, &semaphoreInfo, NULL,
										&r->renderFinishedSemaphores[i]);

				if (ret != VK_SUCCESS) {
					return ret;
				}

				ret = vkCreateFence(r->device, &fenceInfo, NULL,
									&r->inFlightFences[i]);

				if (ret != VK_SUCCESS) {
					return ret;
				}
			}
		}
	}

	return VK_SUCCESS;
}

void vkDestroySwapChain(renderer* r) {
	VK_PROC_ADDR(vkFreeCommandBuffers);
	VK_PROC_ADDR(vkDestroySemaphore);
	VK_PROC_ADDR(vkDestroyFence);
	VK_PROC_ADDR(vkDestroyFramebuffer);
	VK_PROC_ADDR(vkDestroyImageView);
	VK_PROC_ADDR(vkDestroyPipeline);
	VK_PROC_ADDR(vkDestroyPipelineLayout);
	VK_PROC_ADDR(vkDestroyRenderPass);
	VK_PROC_ADDR(vkDestroySwapchainKHR);

	vkFreeCommandBuffers(r->device, r->commandPool, r->swapChainSz,
						 r->commandbuffers);

	for (uint32_t i = 0; i < r->swapChainSz; i++) {
		vkDestroyFence(r->device, r->inFlightFences[i], NULL);
		vkDestroySemaphore(r->device, r->renderFinishedSemaphores[i], NULL);
		vkDestroySemaphore(r->device, r->imageAvailableSemaphores[i], NULL);
		vkDestroyFramebuffer(r->device, r->framebuffers[i], NULL);
		vkDestroyImageView(r->device, r->swapChainImageViews[i], NULL);
	}

	free(r->commandbuffers);
	free(r->inFlightFences);
	free(r->renderFinishedSemaphores);
	free(r->imageAvailableSemaphores);
	free(r->framebuffers);
	free(r->swapChainImageViews);

	vkDestroyPipeline(r->device, r->pipeline, NULL);
	vkDestroyPipelineLayout(r->device, r->pipelineLayout, NULL);
	vkDestroyRenderPass(r->device, r->renderPass, NULL);
	vkDestroySwapchainKHR(r->device, r->swapChain, NULL);
}
