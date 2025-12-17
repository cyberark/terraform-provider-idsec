// Copyright (c) CyberArk
// SPDX-License-Identifier: Apache-2.0

package schemas

type Tuple[A any, B any] struct {
	First  A
	Second B
}
