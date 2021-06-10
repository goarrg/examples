/*
Copyright 2021 The goARRG Authors.

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
	"os"
	"os/exec"
	"syscall"
	"testing"
)

func runCommand(filename string) *exec.Cmd {
	cmd := exec.Command(filename)

	// we need this to send ctrl-c on windows and not close the test
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	return cmd
}

// There is no sig interrupt on windows, so we send it a ctrl-c event
func sigInterrupt(t *testing.T, process *os.Process) {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		t.Fatal(err)
	}
	proc, err := dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		t.Fatal(err)
	}
	ret, _, err := proc.Call(syscall.CTRL_BREAK_EVENT, uintptr(process.Pid))
	if ret == 0 {
		t.Fatal(err)
	}
}
