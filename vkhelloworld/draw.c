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

#include "renderer.h"

void vkDraw(renderer* r) {
	VK_PROC_ADDR(vkWaitForFences);
	VK_PROC_ADDR(vkAcquireNextImageKHR);
	VK_PROC_ADDR(vkResetFences);
	VK_PROC_ADDR(vkQueueSubmit);
	VK_PROC_ADDR(vkQueuePresentKHR);

	vkWaitForFences(r->device, 1, &r->inFlightFences[r->currentFrame], VK_TRUE,
					UINT64_MAX);

	uint32_t imageIndex;
	if (vkAcquireNextImageKHR(r->device, r->swapChain, UINT64_MAX,
							  r->imageAvailableSemaphores[r->currentFrame],
							  VK_NULL_HANDLE, &imageIndex) != VK_SUCCESS) {
		return;
	}

	VkSubmitInfo submitInfo = {};
	submitInfo.sType = VK_STRUCTURE_TYPE_SUBMIT_INFO;

	VkSemaphore waitSemaphores[] = {
		r->imageAvailableSemaphores[r->currentFrame],
	};
	VkPipelineStageFlags waitStages[] = {
		VK_PIPELINE_STAGE_COLOR_ATTACHMENT_OUTPUT_BIT,
	};
	submitInfo.waitSemaphoreCount = 1;
	submitInfo.pWaitSemaphores = waitSemaphores;
	submitInfo.pWaitDstStageMask = waitStages;

	submitInfo.commandBufferCount = 1;
	submitInfo.pCommandBuffers = &r->commandbuffers[imageIndex];

	VkSemaphore signalSemaphores[] = {
		r->renderFinishedSemaphores[r->currentFrame],
	};
	submitInfo.signalSemaphoreCount = 1;
	submitInfo.pSignalSemaphores = signalSemaphores;

	vkResetFences(r->device, 1, &r->inFlightFences[r->currentFrame]);

	if (vkQueueSubmit(r->graphicsQueue, 1, &submitInfo,
					  r->inFlightFences[r->currentFrame]) != VK_SUCCESS) {
		return;
	}

	VkPresentInfoKHR presentInfo = {};
	presentInfo.sType = VK_STRUCTURE_TYPE_PRESENT_INFO_KHR;

	presentInfo.waitSemaphoreCount = 1;
	presentInfo.pWaitSemaphores = signalSemaphores;

	VkSwapchainKHR swapChains[] = {r->swapChain};
	presentInfo.swapchainCount = 1;
	presentInfo.pSwapchains = swapChains;

	presentInfo.pImageIndices = &imageIndex;

	if (vkQueuePresentKHR(r->presentQueue, &presentInfo) != VK_SUCCESS) {
		return;
	}

	r->currentFrame = (r->currentFrame + 1) % r->swapChainSz;
}
