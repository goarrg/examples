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

	goarrg.Install(
		goarrg.Dependencies{
			Target: target,
			SDL:    goarrg.SDLConfig{Install: true, Build: toolchain.BuildRelease},
		},
	)
	vxr.Install(target, toolchain.BuildRelease)

	if golang.ShouldCleanCache() {
		golang.CleanCache()
	}

	if err := toolchain.Run("go", "build", filepath.Join(golang.CallersPackage(packages.NeedFiles).Dir, "..", "..")); err != nil {
		panic(err)
	}
}
