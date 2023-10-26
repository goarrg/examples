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

#include <windows.h>

extern "C" {
void sigInterrupt(uintptr_t pid) {
	HWND h = nullptr;
	do {
		h = FindWindowEx(nullptr, h, nullptr, nullptr);
		DWORD checkProcessID = 0;
		GetWindowThreadProcessId(h, &checkProcessID);
		if (checkProcessID == pid) {
			PostMessage(h, WM_CLOSE, 0, 0);
		}
	} while (h != nullptr);
}
}
