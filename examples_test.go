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
	"goarrg.com/toolchain"
	"goarrg.com/toolchain/cgodep"
	"goarrg.com/toolchain/golang"
)

func TestExamples(t *testing.T) {
	target := toolchain.Target{OS: runtime.GOOS, Arch: runtime.GOARCH}
	cgodep.Install()
	golang.Setup(golang.Config{Target: target})
	debug.IPrintf("Env:\n%s", toolchain.EnvString())

	goarrg.Install(
		goarrg.Dependencies{
			Target:    target,
			SDL:       goarrg.SDLConfig{Install: true, Build: toolchain.BuildRelease},
			VkHeaders: goarrg.VkHeadersConfig{Install: true},
		},
	)

	if golang.ShouldCleanCache() {
		golang.CleanCache()
	}

	files, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}

	tmpDir, err := os.MkdirTemp("", "goarrg")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	for _, f := range files {
		if f.IsDir() {
			switch {
			case f.Name() == "shared":
				continue

			case strings.HasPrefix(f.Name(), "."):
				continue
			}

			t.Run(f.Name(), func(t *testing.T) {
				if strings.HasPrefix(f.Name(), "vk") && !enableVK {
					t.Skip("Vulkan disabled, skipping", f.Name())
				}
				if strings.HasPrefix(f.Name(), "gl") && !enableGL {
					t.Skip("OpenGL disabled, skipping", f.Name())
				}
				if strings.HasPrefix(f.Name(), "vkgl") && (!enableVK || !enableGL) {
					t.Skip("Vulkan and/or OpenGL disabled, skipping", f.Name())
				}

				filename := filepath.Join(tmpDir, f.Name())

				if runtime.GOOS == "windows" {
					filename += ".exe"
				}

				// build
				{
					args := []string{"build", "-o=" + filename}

					if enableDebug {
						args = append(args, "-tags=debug")
					}

					if err := toolchain.RunDir(f.Name(), "go", args...); err != nil {
						t.Fatal(err)
					}
				}

				cmd := runCommand(filename)
				cmd.Dir = filepath.Join(".", f.Name())
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
			})
		}
	}
}
