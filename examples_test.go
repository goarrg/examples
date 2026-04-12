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

package examples

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"goarrg.com/debug"
	goarrg "goarrg.com/make"
	vxr "goarrg.com/rhi/vxr/make"
	"goarrg.com/toolchain"
	"goarrg.com/toolchain/cgodep"
	"goarrg.com/toolchain/golang"
)

func TestExamples(t *testing.T) {
	target := toolchain.Target{OS: runtime.GOOS, Arch: runtime.GOARCH}
	cgodep.Install()
	golang.Setup(golang.Config{Target: target})
	debug.IPrintf("Env:\n%s", toolchain.EnvString())

	buildOptions := goarrg.BuildOptions{}

	if enableDebug {
		buildOptions.Build = toolchain.BuildDebug
	} else {
		buildOptions.Build = toolchain.BuildRelease
	}

	buildTags := goarrg.Install(
		goarrg.Config{
			Target: target,
			Dependencies: goarrg.Dependencies{
				SDL: goarrg.SDLConfig{Install: true, ForceStatic: true, Build: toolchain.BuildRelease},
				Vulkan: goarrg.VulkanDependencies{
					InstallHeaders: true,
				},
			},
			BuildOptions: buildOptions,
		},
	)

	if enableVK {
		if buildTags != "" {
			buildTags += ","
		}
		buildTags += vxr.Install(vxr.Config{
			Target: target,
			BuildOptions: vxr.BuildOptions{
				Build: toolchain.BuildRelease,
				Disable: vxr.DisableFeatures{
					ShaderCompiler: true,
				},
			},
		})
	}

	if golang.ShouldCleanCache() {
		golang.CleanCache()
	}

	tmpDir, err := os.MkdirTemp("", "goarrg")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	exec := func(rootdir string, build func(string) string) {
		t.Helper()

		filename := build(tmpDir)

		cmd := runCommand(filename)
		cmd.Dir = filepath.Join(rootdir)
		cmd.Stdout = os.Stdout

		stderr, err := cmd.StderrPipe()
		if err != nil {
			t.Fatal(err)
		}

		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}

		errc := make(chan error, 1)
		s := bufio.NewScanner(stderr)
		for s.Scan() {
			if strings.Contains(s.Text(), "Engine Init took") {
				// send quit signal but keep the scanner flushing to capture output
				go func() {
					defer close(errc)
					time.Sleep(time.Second)
					err := sigInterrupt(cmd.Process)
					if err != nil {
						errc <- err
					}
				}()
			}
			fmt.Fprintln(os.Stderr, s.Text())
		}
		if err := <-errc; err != nil {
			t.Fatal(err)
		}
		if err := cmd.Wait(); err != nil {
			t.Fatal(err)
		}
	}

	// test opengl
	if enableGL {
		files, err := os.ReadDir("./gl")
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if f.IsDir() {
				switch {
				case f.Name() == "shared":
					continue

				case strings.HasPrefix(f.Name(), "."):
					continue
				}

				t.Run(f.Name(), func(t *testing.T) {
					if strings.HasPrefix(f.Name(), "vkgl") && (!enableVK) {
						t.Skip("Vulkan disabled, skipping", f.Name())
					}

					roodir := filepath.Join("gl", f.Name())
					exec(roodir, func(dir string) string {
						filename := filepath.Join(dir, f.Name())
						args := []string{"build", "-tags=" + buildTags, "-o=" + filename}
						if err := toolchain.RunDir(roodir, "go", args...); err != nil {
							t.Fatal(err)
						}
						return filename
					})
				})
			}
		}
	}

	// test vxr
	if enableVK {
		files, err := os.ReadDir("./vk/vxr")
		if err != nil {
			panic(err)
		}
		for _, f := range files {
			if f.IsDir() {
				switch {
				case f.Name() == "shared":
					continue

				case strings.HasPrefix(f.Name(), "."):
					continue
				}

				t.Run(f.Name(), func(t *testing.T) {
					roodir := filepath.Join("vk", "vxr", f.Name())
					exec(roodir, func(dir string) string {
						filename := filepath.Join(dir, f.Name())
						args := []string{"run", "goarrg.com/examples/vk/vxr/" + f.Name() + "/cmd/make"}
						if err := toolchain.RunDir(dir, "go", args...); err != nil {
							t.Fatal(err)
						}
						return filename
					})
				})
			}
		}
	}
}
