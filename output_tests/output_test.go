/* SPDX-License-Identifier: Apache-2.0 */
/* Copyright(c) 2019 Wind River Systems, Inc. */

package output_tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/gofuzz"

	"github.com/wind-river/deepequal-gen/output_tests/aliases"
	"github.com/wind-river/deepequal-gen/output_tests/builtins"
	"github.com/wind-river/deepequal-gen/output_tests/maps"
	"github.com/wind-river/deepequal-gen/output_tests/pointer"
	"github.com/wind-river/deepequal-gen/output_tests/slices"
	"github.com/wind-river/deepequal-gen/output_tests/structs"
)

func TestWithValueFuzzer(t *testing.T) {
	tests := []interface{}{
		aliases.Ttest{},
		builtins.Ttest{},
		maps.Ttest{},
		pointer.Ttest{},
		slices.Ttest{},
		structs.Ttest{},
	}

	fuzzer := fuzz.New()
	fuzzer.NilChance(0.5)
	fuzzer.NumElements(0, 2)

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test), func(t *testing.T) {
			N := 1000
			for i := 0; i < N; i++ {
				original := reflect.New(reflect.TypeOf(test)).Interface()

				fuzzer.Fuzz(original)

				reflectCopy := ReflectDeepCopy(original)

				if !reflect.DeepEqual(original, reflectCopy) {
					t.Errorf("original and reflectCopy are different:\n\n  original = %s\n\n  jsonCopy = %s", spew.Sdump(original), spew.Sdump(reflectCopy))
				}
			}
		})
	}
}

func BenchmarkReflectDeepEqual(b *testing.B) {
	fourtytwo := "fourtytwo"

	tests := []interface{}{
		maps.Ttest{
			Byte:      map[string]byte{"0": 42, "1": 42, "3": 42},
			Int16:     map[string]int16{"0": 42, "1": 42, "3": 42},
			Int32:     map[string]int32{"0": 42, "1": 42, "3": 42},
			Int64:     map[string]int64{"0": 42, "1": 42, "3": 42},
			Uint8:     map[string]uint8{"0": 42, "1": 42, "3": 42},
			Uint16:    map[string]uint16{"0": 42, "1": 42, "3": 42},
			Uint32:    map[string]uint32{"0": 42, "1": 42, "3": 42},
			Uint64:    map[string]uint64{"0": 42, "1": 42, "3": 42},
			Float32:   map[string]float32{"0": 42.0, "1": 42.0, "3": 42.0},
			Float64:   map[string]float64{"0": 42, "1": 42, "3": 42},
			String:    map[string]string{"0": "fourtytwo", "1": "fourtytwo", "3": "fourtytwo"},
			StringPtr: map[string]*string{"0": &fourtytwo, "1": &fourtytwo, "3": &fourtytwo},
			Struct:    map[string]maps.Ttest{"0": {}, "1": {Byte: map[string]byte{"0": 42, "1": 42, "3": 42}}},
			StructPtr: map[string]*maps.Ttest{"0": nil, "1": {}, "2": {Byte: map[string]byte{"0": 42, "1": 42, "3": 42}}},
		},
		slices.Ttest{
			Byte:      []byte{42, 42, 42},
			Int16:     []int16{42, 42, 42},
			Int32:     []int32{42, 42, 42},
			Int64:     []int64{42, 42, 42},
			Uint8:     []uint8{42, 42, 42},
			Uint16:    []uint16{42, 42, 42},
			Uint32:    []uint32{42, 42, 42},
			Uint64:    []uint64{42, 42, 42},
			Float32:   []float32{42.0, 42.0, 42.0},
			Float64:   []float64{42, 42, 42},
			String:    []string{"fourtytwo", "fourtytwo", "fourtytwo"},
			StringPtr: []*string{&fourtytwo, &fourtytwo, &fourtytwo},
			Struct:    []slices.Ttest{{}, {Byte: []byte{42, 42, 42}}},
			StructPtr: []*slices.Ttest{nil, {}, {Byte: []byte{42, 42, 42}}},
		},
		pointer.Ttest{
			Builtin: &fourtytwo,
			Struct: &pointer.Ttest{
				Builtin: &fourtytwo,
			},
		},
	}

	fuzzer := fuzz.New()
	fuzzer.NilChance(0.5)
	fuzzer.NumElements(0, 2)

	for _, test := range tests {
		b.Run(fmt.Sprintf("%T", test), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				switch t := test.(type) {
				case maps.Ttest:
					t.DeepEqual(&t)
				case slices.Ttest:
					t.DeepEqual(&t)
				case pointer.Ttest:
					t.DeepEqual(&t)
				default:
					b.Fatalf("missing type case in switch for %T", t)
				}
			}
		})
	}
}
