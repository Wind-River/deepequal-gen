/*
SPDX-License-Identifier: Apache-2.0
Copyright 2016 The Kubernetes Authors.
*/

package wholepkg

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
)

func TestDeepEqualPrimitives(t *testing.T) {
	x := StructPrimitives{}
	y := StructPrimitives{}

	if !reflect.DeepEqual(&x, &y) {
		t.Errorf("objects should be equal to start, but are not")
	}

	if reflect.DeepEqual(&x, &y) != x.DeepEqual(&y) {
		t.Errorf("reflect.DeepEqual and DeepEqual do not agree")
	}

	fuzzer := fuzz.New()
	fuzzer.Fuzz(&x)
	fuzzer.Fuzz(&y)

	if reflect.DeepEqual(&x, &y) {
		t.Errorf("objects should not be equal, but are")
	}

	if reflect.DeepEqual(&x, &y) != x.DeepEqual(&y) {
		t.Errorf("reflect.DeepEqual and DeepEqual do not agree")
	}
}
