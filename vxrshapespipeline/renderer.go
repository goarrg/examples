//go:build !goarrg_disable_vk
// +build !goarrg_disable_vk

//go:generate go run goarrg.com/rhi/vxr/cmd/vxrc -dir=./shaders -generator=go -skip-metadata main.frag
//go:generate go run goarrg.com/rhi/vxr/cmd/vxrc -dir=./shaders -generator=go -skip-metadata line.frag
//go:generate go run goarrg.com/rhi/vxr/cmd/vxrc -dir=./shaders -generator=go -skip-metadata square.frag

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
	"fmt"
	"image"
	"math"
	"time"
	"unsafe"

	"goarrg.com"
	"goarrg.com/gmath"
	"goarrg.com/rhi/vxr"
	"goarrg.com/rhi/vxr/shapes"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type renderer struct {
	renderFinishedSemaphore *vxr.TimelineSemaphore
	fragShader              *vxr.GraphicsShaderPipeline
	shapesPipeline          *shapes.Pipeline2D

	lineWidth               float32
	lineFragShader          *vxr.GraphicsShaderPipeline
	lineShapesPipeline      *shapes.Pipeline2DLine
	lineStripShapesPipeline *shapes.Pipeline2DLineStrip

	squareFragShader     *vxr.GraphicsShaderPipeline
	squareSampler        *vxr.Sampler
	squareTexture        *vxr.DeviceColorImage
	squareDescriptorSet  *vxr.DescriptorSet
	squareShapesPipeline *shapes.Pipeline2D
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
		r.shapesPipeline = shapes.NewPipeline2DRegularNGonStar(fl, nil, 4)
	}
	{
		r.lineWidth = min(8, vxr.DeviceProperties().Limits.LineWidth.Max)
		fs, fl := vxrcLoad_line_frag()
		r.lineFragShader = vxr.NewGraphicsShaderPipeline(vxr.NewPipelineLayout(
			vxr.PipelineLayoutCreateInfo{
				ShaderLayout: fl, ShaderStage: vxr.ShaderStageFragment,
			},
		), fs, fl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
		r.lineShapesPipeline = shapes.NewPipeline2DLine(fl, nil)
		r.lineStripShapesPipeline = shapes.NewPipeline2DLineStrip(fl, nil)
	}
	{
		const maxTextureCount = 1
		fs, fl := vxrcLoad_square_frag()
		layout := vxr.NewPipelineLayout(
			vxr.PipelineLayoutCreateInfo{
				ShaderLayout: fl, ShaderStage: vxr.ShaderStageFragment,
				SpecConstants: []uint32{maxTextureCount},
			},
		)
		r.squareFragShader = vxr.NewGraphicsShaderPipeline(layout, fs, fl.EntryPoints["main"], vxr.GraphicsShaderPipelineCreateInfo{})
		r.squareShapesPipeline = shapes.NewPipeline2DSquare(fl, []uint32{maxTextureCount})

		{
			f, err := opentype.Parse(goitalic.TTF)
			if err != nil {
				panic(fmt.Sprintf("Parse: %v", err))
			}
			face, err := opentype.NewFace(f, &opentype.FaceOptions{
				Size:    64,
				DPI:     72,
				Hinting: font.HintingNone,
			})
			if err != nil {
				panic(fmt.Sprintf("NewFace: %v", err))
			}

			bounds, _ := font.BoundString(face, "Hello World")
			extent := gmath.Extent2i32{
				X: int32(bounds.Max.X.Ceil()),
				Y: int32(face.Metrics().Height.Ceil()),
			}
			dst := image.NewRGBA(image.Rect(0, 0, int(extent.X), int(extent.Y)))
			d := font.Drawer{
				Dst:  dst,
				Src:  image.White,
				Face: face,
				Dot:  fixed.P(0, face.Metrics().Ascent.Ceil()),
			}
			d.DrawString("Hello World")

			r.squareTexture = vxr.NewColorImage("hello World", vxr.FORMAT_R8G8B8A8_SRGB, vxr.ImageCreateInfo{
				Usage:          vxr.ImageUsageSampled | vxr.ImageUsageTransferDst,
				Extent:         gmath.Extent3i32{X: extent.X, Y: extent.Y, Z: 1},
				NumMipLevels:   1,
				NumArrayLayers: 1,
			})

			frame := vxr.FrameBegin()
			b := frame.NewHostScratchBuffer("upload", r.squareTexture.BufferSize(), vxr.BufferUsageTransferSrc)
			b.HostWrite(0,
				unsafe.Slice(
					(*byte)(unsafe.Pointer(unsafe.SliceData(dst.Pix))), uint64(unsafe.Sizeof(dst.Pix[0]))*uint64(len(dst.Pix)),
				),
			)
			cb := frame.NewSingleUseCommandBuffer("upload")
			cb.ImageBarrier(vxr.ImageBarrier{
				Image:  r.squareTexture,
				Aspect: r.squareTexture.Aspect(),
				Src: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageNone,
					Access: vxr.AccessFlagNone,
					Layout: vxr.ImageLayoutUndefined,
				},
				Dst: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageTransfer,
					Access: vxr.AccessFlagMemoryWrite,
					Layout: vxr.ImageLayoutTransferDst,
				},
				Range: vxr.ImageSubresourceRange{
					NumMipLevels:   1,
					NumArrayLayers: 1,
				},
			})
			cb.CopyBufferToImage(b, r.squareTexture, vxr.ImageLayoutTransferDst, []vxr.BufferImageCopyRegion{
				{
					ImageSubresource: vxr.ImageSubresourceLayers{NumArrayLayers: 1},
					ImageExtent:      r.squareTexture.Extent(),
				},
			})
			cb.ImageBarrier(vxr.ImageBarrier{
				Image:  r.squareTexture,
				Aspect: r.squareTexture.Aspect(),
				Src: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageTransfer,
					Access: vxr.AccessFlagMemoryWrite,
					Layout: vxr.ImageLayoutTransferDst,
				},
				Dst: vxr.ImageBarrierInfo{
					Stage:  vxr.PipelineStageFragmentShader,
					Access: vxr.AccessFlagMemoryRead,
					Layout: vxr.ImageLayoutReadOnlyOptimal,
				},
				Range: vxr.ImageSubresourceRange{
					NumMipLevels:   1,
					NumArrayLayers: 1,
				},
			})
			cb.Submit(nil, []vxr.SemaphoreSignalInfo{{
				Semaphore: r.renderFinishedSemaphore,
				Stage:     vxr.PipelineStageTransfer,
			}})
			frame.EndWithWaiter(r.renderFinishedSemaphore.WaiterForPendingValue())

			r.squareSampler = vxr.NewSampler("main", vxr.SamplerCreateInfo{
				MagFilter:  vxr.SamplerFilterNearest,
				MinFilter:  vxr.SamplerFilterNearest,
				MipMapMode: vxr.SamplerMipMapModeNearest,
				BorderMode: vxr.SamplerAddressModeClampToEdge,
			})
			r.squareDescriptorSet = layout.NewDescriptorSet(1)
			r.squareDescriptorSet.Bind(0, 0, vxr.DescriptorCombinedImageSamplerInfo{
				Sampler: r.squareSampler,
				Image:   r.squareTexture,
				Layout:  vxr.ImageLayoutReadOnlyOptimal,
			})
		}
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

		r.squareShapesPipeline.Draw(frame, cb, r.squareFragShader, viewport,
			vxr.DrawParameters{
				DescriptorSets: []*vxr.DescriptorSet{r.squareDescriptorSet},
			},
			shapes.InstanceData2D{
				Transform: shapes.Transform2D{
					Pos:  gmath.Point2f32{Y: 300},
					Size: gmath.Vector2f32{X: 240, Y: 128},
				},
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
	r.renderFinishedSemaphore.Wait()
	r.renderFinishedSemaphore.Destroy()

	r.fragShader.Destroy()
	r.shapesPipeline.Destroy()

	r.lineFragShader.Destroy()
	r.lineShapesPipeline.Destroy()
	r.lineStripShapesPipeline.Destroy()

	r.squareFragShader.Destroy()
	r.squareSampler.Destroy()
	r.squareTexture.Destroy()
	r.squareDescriptorSet.Destroy()
	r.squareShapesPipeline.Destroy()

	shapes.Destroy()
	vxr.Destroy()
}
