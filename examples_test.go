package examples

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "goarrg.com/cmd/goarrg"
)

func TestExamples(t *testing.T) {
	files, err := ioutil.ReadDir(".")

	if err != nil {
		panic(err)
	}

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

				cmd := exec.Command("go", "run", "goarrg.com/cmd/goarrg", "run")
				cmd.Dir = filepath.Join(".", f.Name())
				cmd.Stderr = os.Stderr
				cmd.Stdout = os.Stdout

				if err := cmd.Run(); err != nil {
					t.Fatal(err)
				}

				t.Run("debug", func(t *testing.T) {
					cmd := exec.Command("go", "run", "goarrg.com/cmd/goarrg", "run", "--", "-tags=debug")
					cmd.Dir = filepath.Join(".", f.Name())
					cmd.Stderr = os.Stderr
					cmd.Stdout = os.Stdout

					if err := cmd.Run(); err != nil {
						t.Fatal(err)
					}
				})
			})
		}
	}
}
