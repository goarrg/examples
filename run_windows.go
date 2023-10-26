/*
Copyright 2023 The goARRG Authors.

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

/*
Copyright 2023 The goARRG Authors.

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

/*
	#include <stdint.h>
	extern void sigInterrupt(uintptr_t);
*/
import "C"

import (
	"os"
	"os/exec"
	"syscall"
)

func runCommand(filename string) *exec.Cmd {
	cmd := exec.Command(filename)
	return cmd
}

// There is no sig interrupt on windows, so we send it a WM_CLOSE event
func sigInterrupt(process *os.Process) error {
	C.sigInterrupt(C.uintptr_t(process.Pid))
	return syscall.GetLastError()
}
