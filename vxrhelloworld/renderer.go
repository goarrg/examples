//go:build !goarrg_disable_vk
// +build !goarrg_disable_vk

/*
Copyright 2025 The goARRG Authors.

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
	"path/filepath"
	"time"

	"goarrg.com"
	"goarrg.com/asset"
	"goarrg.com/gmath"
	"goarrg.com/rhi/vxr"
	"goarrg.com/toolchain/golang"
	"golang.org/x/tools/go/packages"
)

type renderer struct {
	triangleLayout          *vxr.PipelineLayout
	trianglePipeline        vxr.GraphicsPipelineLibrary
	renderFinishedSemaphore *vxr.TimelineSemaphore
}

func (r *renderer) VkConfig() goarrg.VkConfig {
	return vxr.VkConfig()
}

func (r *renderer) VkInit(platform goarrg.PlatformInterface, vkInstance goarrg.VkInstance) error {
	filesystem := asset.DirFS(filepath.Join(golang.CallersPackage(packages.NeedFiles).Dir, "./shaders"))
	vxr.Init(platform, vkInstance, vxr.Config{
		MaxFramesInFlight:      2,
		DescriptorPoolBankSize: 1,
	})

	vi := vxr.NewVertexInputPipeline(vxr.VertexInputPipelineCreateInfo{
		Topology: vxr.VertexTopologyTriangleList,
	})

	vs, vl, _ := vxr.CompileShader(filesystem, "main.vert")
	fs, fl, _ := vxr.CompileShader(filesystem, "main.frag")
	r.triangleLayout = vxr.NewPipelineLayout(
		vxr.PipelineLayoutCreateInfo{
			ShaderLayout: vl, ShaderStage: vxr.ShaderStageVertex,
		},
		vxr.PipelineLayoutCreateInfo{
			ShaderLayout: fl, ShaderStage: vxr.ShaderStageFragment,
		},
	)
	vp := vxr.NewGraphicsShaderPipeline(r.triangleLayout, vs, vl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
	fp := vxr.NewGraphicsShaderPipeline(r.triangleLayout, fs, fl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
	r.trianglePipeline = vxr.GraphicsPipelineLibrary{
		Layout: r.triangleLayout, VertexInput: vi, VertexShader: vp, FragmentShader: fp,
	}
	r.renderFinishedSemaphore = vxr.NewTimelineSemaphore("renderFinishedSemaphore")
	return nil
}

func (r *renderer) Draw() float64 {
	frame := vxr.FrameBegin()

	if frame.Surface() == nil {
		frame.Cancel()
		time.Sleep(time.Millisecond * 50)
		return 0
	}

	cb := frame.NewSingleUseCommandBuffer("main")
	{
		cb.ImageBarrier(
			vxr.ImageBarrier{
				Image: frame.Surface(),
				Src: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageRenderAttachmentWrite,
					Access: vxr.AccessFlagNone,
					Layout: vxr.ImageLayoutUndefined,
				},
				Dst: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageRenderAttachmentWrite,
					Access: vxr.AccessFlagMemoryWrite,
					Layout: vxr.ImageLayoutAttachmentOptimal,
				},
				Range: vxr.ImageSubresourceRange{BaseMipLevel: 0, NumMipLevels: 1, BaseArrayLayer: 0, NumArrayLayers: 1},
			},
		)

		cb.RenderPassBegin("triangle", gmath.Recti32{W: frame.Surface().Extent().X, H: frame.Surface().Extent().Y},
			vxr.RenderAttachments{
				Color: []vxr.RenderColorAttachment{
					{
						Image:               frame.Surface(),
						Layout:              vxr.ImageLayoutAttachmentOptimal,
						LoadOp:              vxr.RenderAttachmentLoadOpClear,
						StoreOp:             vxr.RenderAttachmentStoreOpStore,
						ColorBlendEnable:    true,
						ColorBlendEquation:  vxr.RenderColorAttachmentBlendPremultipliedAlpha(),
						ColorComponentFlags: frame.Surface().Format().ColorComponentFlags(),
					},
				},
			})

		cb.Draw(r.trianglePipeline, vxr.DrawInfo{
			VertexCount:   3,
			InstanceCount: 1,
		})

		cb.RenderPassEnd()

		cb.ImageBarrier(
			vxr.ImageBarrier{
				Image: frame.Surface(),
				Src: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageRenderAttachmentWrite,
					Access: vxr.AccessFlagMemoryWrite,
					Layout: vxr.ImageLayoutAttachmentOptimal,
				},
				Dst: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageRenderAttachmentWrite,
					Access: vxr.AccessFlagNone,
					Layout: vxr.ImageLayoutPresent,
				},
				Range: vxr.ImageSubresourceRange{BaseMipLevel: 0, NumMipLevels: 1, BaseArrayLayer: 0, NumArrayLayers: 1},
			},
		)

		cb.Submit(
			[]vxr.SemaphoreWaitInfo{
				{Semaphore: frame.Surface(), Stage: vxr.PipelineStageRenderAttachmentWrite},
			},
			[]vxr.SemaphoreSignalInfo{
				{Semaphore: frame.Surface(), Stage: vxr.PipelineStageRenderAttachmentWrite},
				{Semaphore: r.renderFinishedSemaphore, Stage: vxr.PipelineStageRenderAttachmentWrite},
			},
		)
	}

	frame.EndWithWaiter(r.renderFinishedSemaphore.WaiterForPendingValue())
	return 0
}

func (r *renderer) Resize(w int, h int) {
	vxr.Resize(w, h)
}

func (r *renderer) Destroy() {
	r.renderFinishedSemaphore.Destroy()
	vxr.Destroy()
}
