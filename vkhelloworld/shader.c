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

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "renderer.h"

extern void* assetLoad(char*, size_t*);
extern void assetFree(char*);

VkResult vkShaderLoad(renderer* r, const char* file, VkShaderModule* module) {
	VK_PROC_ADDR_ERROR(vkCreateShaderModule);

	size_t sz;
	void* bytes = assetLoad((char*)file, &sz);

	VkShaderModuleCreateInfo createInfo = {};
	createInfo.sType = VK_STRUCTURE_TYPE_SHADER_MODULE_CREATE_INFO;
	createInfo.codeSize = sz;
	createInfo.pCode = (const uint32_t*)(bytes);

	VkResult ret = vkCreateShaderModule(r->device, &createInfo, NULL, module);

	assetFree((char*)file);

	return ret;
}
