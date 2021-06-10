//+build !disable_vk,amd64

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

package main

/*
	#include <stdint.h>
*/
import "C"
import (
	"goarrg.com/asset"
)

var assets = map[string]asset.Asset{}

//export assetLoad
func assetLoad(name *C.char, sz *C.size_t) C.uintptr_t {
	a, err := asset.Load(C.GoString(name))

	if err != nil {
		panic(err)
	}

	assets[a.Filename()] = a
	*sz = (C.size_t)(a.Size())

	return C.uintptr_t(a.Uintptr())
}

//export assetFree
func assetFree(name *C.char) {
	delete(assets, C.GoString(name))
}
