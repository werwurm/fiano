// Copyright 2017-2023 the LinuxBoot Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate manifestcodegen

package bootpolicy

type PM struct {
	StructInfo `id:"__PMDA__" version:"0x10"`
	Data       []byte `json:"pcData"`
}
