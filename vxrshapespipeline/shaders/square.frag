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

#version 450
#extension GL_ARB_separate_shader_objects : enable
#extension GL_EXT_nonuniform_qualifier : enable
#pragma shader_stage(fragment)

layout(constant_id = 0) const int maxTextureCount = 1;

layout(location = 0) in flat uint instanceID;
layout(location = 1) in vec2 uv;

layout(location = 0) out vec4 outColor;

// set 0 is reserved for vertex shader
layout(set = 1, binding = 0) uniform sampler2D textures[maxTextureCount];

void main() {
	outColor = texture(textures[nonuniformEXT(instanceID)], uv);
}
