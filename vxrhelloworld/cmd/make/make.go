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
	"os"
	"path/filepath"
	"runtime"

	"goarrg.com/debug"
	goarrg "goarrg.com/make"
	vxr "goarrg.com/rhi/vxr/make"
	"goarrg.com/toolchain"
	"goarrg.com/toolchain/cgodep"
	"goarrg.com/toolchain/golang"
	"golang.org/x/tools/go/packages"
)

func main() {
	target := toolchain.Target{OS: runtime.GOOS, Arch: runtime.GOARCH}
	cgodep.Install()
	golang.Setup(golang.Config{Target: target})
	debug.IPrintf("Env:\n%s", toolchain.EnvString())

	buildTags := goarrg.Install(
		goarrg.Config{
			Target: target,
			Dependencies: goarrg.Dependencies{
				SDL:       goarrg.SDLConfig{Install: true, Build: toolchain.BuildRelease},
				VkHeaders: goarrg.VkHeadersConfig{Install: true},
			},
		},
	)
	if buildTags != "" {
		buildTags += ","
	}

	if len(os.Args) == 1 {
		buildTags += vxr.Install(vxr.Config{
			Target: target,
			BuildOptions: vxr.BuildOptions{
				Build: toolchain.BuildRelease,
				Disable: vxr.DisableFeatures{
					ShaderCompiler: true,
				},
			},
		})

		if golang.ShouldCleanCache() {
			golang.CleanCache()
		}
		if err := toolchain.Run("go", "build", "-tags="+buildTags, filepath.Join(golang.CallersPackage(packages.NeedFiles).Dir, "..", "..")); err != nil {
			panic(err)
		}
	} else if (len(os.Args) == 2) && (os.Args[1] == "generate") {
		// generate requires the shader compiler, duh, so we need to recompile vxr
		// but since generate does not get called all the time and vxr compiles fast
		// this isn't too bad
		buildTags += vxr.Install(vxr.Config{
			Target: target,
			BuildOptions: vxr.BuildOptions{
				Build: toolchain.BuildRelease,
			},
		})

		if golang.ShouldCleanCache() {
			golang.CleanCache()
		}
		if err := toolchain.Run("go", "generate", "-tags="+buildTags, filepath.Join(golang.CallersPackage(packages.NeedFiles).Dir, "..", "..")); err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Sprintf("Invalid args: %q", os.Args[1:]))
	}
}
