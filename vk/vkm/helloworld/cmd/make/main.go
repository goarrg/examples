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
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"goarrg.com/debug"
	vkm "goarrg.com/lib/vkm/make"
	goarrg "goarrg.com/make"
	"goarrg.com/toolchain"
	"goarrg.com/toolchain/cc"
	"goarrg.com/toolchain/cgodep"
	"goarrg.com/toolchain/golang"
	"golang.org/x/tools/go/packages"
)

func setup(target toolchain.Target, build toolchain.Build) {
	cgodep.Install()
	cc.Setup(cc.Config{Compiler: cc.CompilerClang, Target: target}, build)
	debug.IPrintf("Env:\n%s", toolchain.EnvString())

	goarrg.Install(
		goarrg.Config{
			Target: target,
			Dependencies: goarrg.Dependencies{
				SDL: goarrg.SDLConfig{Install: true, ForceStatic: true, Build: toolchain.BuildRelease},
				Vulkan: goarrg.VulkanDependencies{
					InstallHeaders: true,
					InstallDocs:    true,
					Shaderc:        goarrg.ShadercConfig{Install: true, ForceStatic: true, Build: toolchain.BuildRelease},
				},
			},
		},
	)
	vkm.Install(vkm.Config{
		Target: target,
		BuildOptions: vkm.BuildOptions{
			Build: build,
		},
	})
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	srcDir := filepath.Join(golang.CallersPackage(packages.LoadFiles).Dir, "..", "..")
	target := toolchain.Target{OS: runtime.GOOS, Arch: runtime.GOARCH}
	setup(target, toolchain.BuildDebug)

	var cFlags, ldFlags []string
	{
		var err error
		deps := []string{"sdl3", "vulkan-headers", "vkm"}
		cFlags, err = cgodep.Resolve(target, cgodep.ResolveCFlags, deps...)
		if err != nil {
			panic(err)
		}
		ldFlags, err = cgodep.Resolve(target, cgodep.ResolveLDFlags|cgodep.ResolveStaticFlags, deps...)
		if err != nil {
			panic(err)
		}
		cFlags = append(cFlags, "-glldb",
			"-Werror=vla", "-Wno-unknown-pragmas", "-Wno-missing-field-initializers", "-Wno-format-security",
		)
		ldFlags = append(ldFlags, "-glldb", "-fuse-ld=lld", "-Wl,-rpath", "-Wl,$ORIGIN")
	}
	buildFlags := cc.BuildFlags{
		CFlags:   append(cFlags, strings.Split(toolchain.EnvGet("CGO_CFLAGS"), " ")...),
		CXXFlags: append(cFlags, strings.Split(toolchain.EnvGet("CGO_CXXFLAGS"), " ")...),
	}
	buildFlags.CFlags = append(buildFlags.CFlags,
		"-std=c17",
	)
	buildFlags.CXXFlags = append(buildFlags.CXXFlags,
		"-std=c++20",
	)
	buildOptions := cc.BuildOptions{
		Type:   cc.BuildTypeExecuteable,
		Target: target,
		CommandOnlyFlags: cc.BuildFlags{
			CFlags:   []string{"-Wall", "-Wextra", "-Wpedantic"},
			CXXFlags: []string{"-Wall", "-Wextra", "-Wpedantic"},
		},
		BuildFlags: buildFlags,
		LDFlags:    append(ldFlags, strings.Split(toolchain.EnvGet("CGO_LDFLAGS"), " ")...),
	}
	buildDir, err := os.MkdirTemp("", "vkm")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(buildDir)
	jsonCmds, err := cc.BuildDir(srcDir, buildDir,
		filepath.Join(wd, "helloworld"+buildOptions.Type.FileExt(target.OS)), buildOptions)
	if err != nil {
		panic(err)
	}
	{
		jOut, err := json.Marshal(jsonCmds)
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "compile_commands.json"), jOut, 0o655); err != nil &&
			!strings.Contains(srcDir, filepath.Join("pkg", "mod")) {
			panic(err)
		}
	}
}
