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
	"time"

	"goarrg.com"
	"goarrg.com/gmath"
	"goarrg.com/gmath/color"
	"goarrg.com/rhi/vxr"
	"goarrg.com/rhi/vxr/shapes"
)

type renderer struct {
	shapesCB                shapes.CommandBuffer2D
	renderFinishedSemaphore *vxr.TimelineSemaphore
	rot                     float32
	lastTime                time.Time
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
	shapes.Init(shapes.Config{})
	r.renderFinishedSemaphore = vxr.NewTimelineSemaphore("renderFinishedSemaphore")
	r.lastTime = time.Now()
	return nil
}

func (r *renderer) Draw() float64 {
	frame := vxr.FrameBegin()

	delta := time.Since(r.lastTime).Seconds()
	r.lastTime = time.Now()
	r.rot += (360 / 4) * float32(delta)
	for r.rot >= 360 {
		r.rot -= 360
	}

	{
		rowOffset := 0
		r.shapesCB.Begin()

		for _, o := range []shapes.TransformOrder{shapes.TransformTRS, shapes.TransformTSR} {
			for col := range uint32(5) {
				r.shapesCB.DrawSquare(
					shapes.Transform2D{
						Pos: gmath.Point2f32{
							X: 8*float32(col*2) + 64*float32(col*2),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
				)
				r.shapesCB.DrawRegularNGon(col+3,
					shapes.Transform2D{
						Rot: gmath.DegToRad(r.rot),
						Pos: gmath.Point2f32{
							X: 8*float32(col*2) + 64*float32(col*2),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
						TransformOrder:   o,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
				)
				r.shapesCB.DrawSquare(
					shapes.Transform2D{
						Pos: gmath.Point2f32{
							X: 8*float32(col*2+1) + 64*float32(col*2+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
				)
				r.shapesCB.DrawRegularNGon(col+3,
					shapes.Transform2D{
						Rot: gmath.DegToRad(r.rot),
						Pos: gmath.Point2f32{
							X: 8*float32(col*2+1) + 64*float32(col*2+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 32},
						TranslationPivot: shapes.PivotTopLeft,
						TransformOrder:   o,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
				)
			}
			rowOffset++
		}

		for _, o := range []shapes.TransformOrder{shapes.TransformTRS, shapes.TransformTSR} {
			for col := range uint32(5) {
				r.shapesCB.DrawSquare(
					shapes.Transform2D{
						Pos: gmath.Point2f32{
							X: 8*float32(col*2) + 64*float32(col*2),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
				)
				r.shapesCB.DrawRegularNGonStar(col+4, 0.5,
					shapes.Transform2D{
						Rot: gmath.DegToRad(r.rot),
						Pos: gmath.Point2f32{
							X: 8*float32(col*2) + 64*float32(col*2),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
						TransformOrder:   o,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
				)
				r.shapesCB.DrawSquare(
					shapes.Transform2D{
						Pos: gmath.Point2f32{
							X: 8*float32(col*2+1) + 64*float32(col*2+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
				)
				r.shapesCB.DrawRegularNGonStar(col+4, 0.5,
					shapes.Transform2D{
						Rot: gmath.DegToRad(r.rot),
						Pos: gmath.Point2f32{
							X: 8*float32(col*2+1) + 64*float32(col*2+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 32},
						TranslationPivot: shapes.PivotTopLeft,
						TransformOrder:   o,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
				)
			}
			rowOffset++
		}

		for i, f := range []func(t shapes.Transform2D, c color.UNorm[uint8]){r.shapesCB.DrawSquare, r.shapesCB.DrawTriangle} {
			r.shapesCB.DrawSquare(
				shapes.Transform2D{
					Pos: gmath.Point2f32{
						X: 8*float32(i*3) + 64*float32(i*3),
						Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
					},
					Size:             gmath.Vector2f32{X: 64, Y: 64},
					TranslationPivot: shapes.PivotTopLeft,
				},
				color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
			)
			f(
				shapes.Transform2D{
					Rot: gmath.DegToRad(r.rot),
					Pos: gmath.Point2f32{
						X: 8*float32(i*3) + 64*float32(i*3),
						Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
					},
					Size:             gmath.Vector2f32{X: 64, Y: 64},
					TranslationPivot: shapes.PivotTopLeft,
				},
				color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
			)
			for j, o := range []shapes.TransformOrder{shapes.TransformTRS, shapes.TransformTSR} {
				r.shapesCB.DrawSquare(
					shapes.Transform2D{
						Pos: gmath.Point2f32{
							X: 8*float32((i*3)+j+1) + 64*float32((i*3)+j+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 64},
						TranslationPivot: shapes.PivotTopLeft,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 255, G: 0, B: 255, A: 255}),
				)
				f(
					shapes.Transform2D{
						Rot: gmath.DegToRad(r.rot),
						Pos: gmath.Point2f32{
							X: 8*float32((i*3)+j+1) + 64*float32((i*3)+j+1),
							Y: (64 * float32(rowOffset)) + (8 * float32(rowOffset+1)),
						},
						Size:             gmath.Vector2f32{X: 64, Y: 32},
						TranslationPivot: shapes.PivotTopLeft,
						TransformOrder:   o,
					},
					color.Convert[color.UNorm[uint8]](color.SRGB[uint8]{R: 0, G: 206, B: 250, A: 255}),
				)
			}
		}

		r.shapesCB.End()
	}

	if frame.Surface() == nil {
		frame.Cancel()
		time.Sleep(time.Millisecond * 50)
		return 0
	}

	cb := frame.NewSingleUseCommandBuffer("main")
	{
		viewport := gmath.Extent2i32{X: frame.Surface().Extent().X, Y: frame.Surface().Extent().Y}
		r.shapesCB.ExecutePrePass(frame, cb, viewport)

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
		r.shapesCB.ExecuteDraw(frame, cb)
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
	r.shapesCB.Destroy()
	shapes.Destroy()
	vxr.Destroy()
}
