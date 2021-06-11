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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExamples(t *testing.T) {
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
		f := f
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
					args := []string{"run", "goarrg.com/cmd/goarrg", "build", "--", "-o=" + filename}

					if enableDebug {
						args = append(args, "-tags=debug")
					}

					cmd := exec.Command("go", args...)
					cmd.Dir = filepath.Join(".", f.Name())
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout
					if err := cmd.Run(); err != nil {
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

				go func() {
					s := bufio.NewScanner(stderr)
					for s.Scan() {
						if strings.Contains(s.Text(), "Engine Init took") {
							// send quit signal but keep the scanner flushing to capture output
							go func() {
								time.Sleep(time.Second)
								sigInterrupt(t, cmd.Process)
							}()
						}
						fmt.Fprintln(os.Stderr, s.Text())
					}
				}()

				if err := cmd.Wait(); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
