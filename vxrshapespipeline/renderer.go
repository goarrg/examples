//go:build !goarrg_disable_vk
// +build !goarrg_disable_vk

//go:generate go run goarrg.com/rhi/vxr/cmd/vxrc -dir=./shaders -generator=go -skip-metadata main.frag
//go:generate go run goarrg.com/rhi/vxr/cmd/vxrc -dir=./shaders -generator=go -skip-metadata line.frag

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
	"math"
	"time"

	"goarrg.com"
	"goarrg.com/gmath"
	"goarrg.com/rhi/vxr"
	"goarrg.com/rhi/vxr/shapes"
)

type renderer struct {
	renderFinishedSemaphore *vxr.TimelineSemaphore
	fragShader              *vxr.GraphicsShaderPipeline
	shapesPipeline          *shapes.Pipeline2D

	lineWidth               float32
	lineFragShader          *vxr.GraphicsShaderPipeline
	lineShapesPipeline      *shapes.Pipeline2DLine
	lineStripShapesPipeline *shapes.Pipeline2DLineStrip
}

func (r *renderer) VkConfig() goarrg.VkConfig {
	return vxr.VkConfig()
}

func (r *renderer) VkInit(platform goarrg.PlatformInterface, vkInstance goarrg.VkInstance) error {
	vxr.InitInstance(platform, vkInstance)
	vxr.InitDevice(vxr.Config{
		RequiredFeatures:       shapes.RequiredVkFeatureStructs(),
		MaxFramesInFlight:      2,
		DescriptorPoolBankSize: 1,
	})
	shapes.Init(platform)
	r.renderFinishedSemaphore = vxr.NewTimelineSemaphore("renderFinishedSemaphore")
	{
		fs, fl := vxrcLoad_main_frag()
		r.fragShader = vxr.NewGraphicsShaderPipeline(vxr.NewPipelineLayout(
			vxr.PipelineLayoutCreateInfo{
				ShaderLayout: fl, ShaderStage: vxr.ShaderStageFragment,
			},
		), fs, fl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
		r.shapesPipeline = shapes.NewPipeline2DRegularNGonStar(fl, 4)
	}
	{
		r.lineWidth = min(8, vxr.DeviceProperties().Limits.LineWidth.Max)
		fs, fl := vxrcLoad_line_frag()
		r.lineFragShader = vxr.NewGraphicsShaderPipeline(vxr.NewPipelineLayout(
			vxr.PipelineLayoutCreateInfo{
				ShaderLayout: fl, ShaderStage: vxr.ShaderStageFragment,
			},
		), fs, fl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
		r.lineShapesPipeline = shapes.NewPipeline2DLine(fl)
		r.lineStripShapesPipeline = shapes.NewPipeline2DLineStrip(fl)
	}
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

		cb.RenderPassBegin("main",
			gmath.Recti32{W: frame.Surface().Extent().X, H: frame.Surface().Extent().Y},
			vxr.RenderParameters{FlipViewport: true},
			vxr.RenderAttachments{
				Color: []vxr.RenderColorAttachment{
					{
						Image:   frame.Surface(),
						Layout:  vxr.ImageLayoutAttachmentOptimal,
						LoadOp:  vxr.RenderAttachmentLoadOpClear,
						StoreOp: vxr.RenderAttachmentStoreOpStore,
						ColorBlend: vxr.RenderColorBlendParameters{
							Enable:         true,
							Equation:       vxr.RenderColorAttachmentBlendPremultipliedAlpha(),
							ComponentFlags: frame.Surface().Format().ColorComponentFlags(),
						},
					},
				},
			})

		viewport := gmath.Extent2i32{X: frame.Surface().Extent().X, Y: frame.Surface().Extent().Y}
		r.shapesPipeline.Draw(frame, cb, r.fragShader, viewport, vxr.DrawParameters{},
			shapes.InstanceData2D{
				Transform: shapes.Transform2D{
					Size: gmath.Vector2f32{X: 128, Y: 128},
				},
				Parameter1: 0.5,
			}, shapes.InstanceData2D{
				Transform: shapes.Transform2D{
					Pos:  gmath.Point2f32{X: 140},
					Size: gmath.Vector2f32{X: 128, Y: 128},
				},
				Parameter1: 0.25,
			})

		r.lineShapesPipeline.Draw(frame, cb, r.lineFragShader, viewport, vxr.DrawParameters{},
			r.lineWidth,
			shapes.InstanceData2DLine{
				P0: gmath.Vector2f32{X: 64, Y: 130},
				P1: gmath.Vector2f32{X: 208, Y: 130},
			},
			shapes.InstanceData2DLine{
				P0: gmath.Vector2f32{X: 208, Y: 130 - (r.lineWidth * 0.5)},
				P1: gmath.Vector2f32{X: 208, Y: 200},
			})

		{
			var points []gmath.Vector2f32
			for x := 0.0; x < 64.0; x++ {
				t := 2 * math.Pi * (x / 64)
				points = append(points, gmath.Vector2f32{X: float32(x) * 4, Y: 240 + float32(math.Sin(t)*32)})
			}
			r.lineStripShapesPipeline.Draw(frame, cb, r.lineFragShader, viewport, vxr.DrawParameters{},
				8, points...)
		}

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
	r.renderFinishedSemaphore.Wait()
	r.renderFinishedSemaphore.Destroy()
	r.fragShader.Destroy()
	r.shapesPipeline.Destroy()
	r.lineFragShader.Destroy()
	r.lineShapesPipeline.Destroy()
	r.lineStripShapesPipeline.Destroy()
	shapes.Destroy()
	vxr.Destroy()
}
