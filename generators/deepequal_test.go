/*
SPDX-License-Identifier: Apache-2.0
Copyright 2016 The Kubernetes Authors.
Copyright 2019 Wind River Systems, Inc.
*/

package generators

import (
	"testing"

	"k8s.io/gengo/types"
)

func Test_deepEqualMethod(t *testing.T) {
	testCases := []struct {
		typ    types.Type
		expect bool
		error  bool
	}{
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				// No DeepEqual method.
				Methods: map[string]*types.Type{},
			},
			expect: false,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// No DeepEqual method.
					"method": {
						Name: types.Name{Package: "pkgname", Name: "func()"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{
								Kind: types.Pointer,
								Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Parameters: []*types.Type{},
							Results:    []*types.Type{},
						},
					},
				},
			},
			expect: false,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Wrong signature (no parameter).
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func()"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{
								Kind: types.Pointer,
								Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Parameters: []*types.Type{},
							Results:    []*types.Type{},
						},
					},
				},
			},
			expect: false,
			error:  true,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Wrong signature (unexpected result).
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func(*pkgname.typename) int"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{
								Kind: types.Pointer,
								Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Parameters: []*types.Type{
								{
									Kind: types.Pointer,
									Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
								},
							},
							Results: []*types.Type{
								{
									Name: types.Name{Name: "int"},
									Kind: types.Builtin,
								},
							},
						},
					},
				},
			},
			expect: false,
			error:  true,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Wrong signature (non-pointer parameter, pointer receiver).
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func(pkgname.typename)"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{
								Kind: types.Pointer,
								Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Parameters: []*types.Type{
								{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Results: []*types.Type{},
						},
					},
				},
			},
			expect: false,
			error:  true,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Wrong signature (non-pointer parameter, non-pointer receiver).
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func(pkgname.typename)"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							Parameters: []*types.Type{
								{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Results: []*types.Type{},
						},
					},
				},
			},
			expect: false,
			error:  true,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Correct signature with non-pointer receiver.
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func(*pkgname.typename)"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							Parameters: []*types.Type{
								{
									Kind: types.Pointer,
									Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
								},
							},
							Results: []*types.Type{
								{
									Name: types.Name{Name: "bool"},
									Kind: types.Builtin,
								},
							},
						},
					},
				},
			},
			expect: true,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// Correct signature with pointer receiver.
					"DeepEqual": {
						Name: types.Name{Package: "pkgname", Name: "func(*pkgname.typename)"},
						Kind: types.Func,
						Signature: &types.Signature{
							Receiver: &types.Type{
								Kind: types.Pointer,
								Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
							},
							Parameters: []*types.Type{
								{
									Kind: types.Pointer,
									Elem: &types.Type{Kind: types.Struct, Name: types.Name{Package: "pkgname", Name: "typename"}},
								},
							},
							Results: []*types.Type{
								{
									Name: types.Name{Name: "bool"},
									Kind: types.Builtin,
								},
							},
						},
					},
				},
			},
			expect: true,
		},
	}

	for i, tc := range testCases {
		r, err := deepEqualMethod(&tc.typ)
		if tc.error && err == nil {
			t.Errorf("case[%d]: expected an error, got none", i)
		} else if !tc.error && err != nil {
			t.Errorf("case[%d]: expected no error, got: %v", i, err)
		} else if !tc.error && (r != nil) != tc.expect {
			t.Errorf("case[%d]: expected result %v, got: %v", i, tc.expect, r)
		}
	}
}

func Test_extractTagParams(t *testing.T) {
	testCases := []struct {
		comments []string
		expect   *enabledTagValue
	}{
		{
			comments: []string{
				"Human comment",
			},
			expect: nil,
		},
		{
			comments: []string{
				"Human comment",
				"+" + tagEnabledName,
			},
			expect: &enabledTagValue{
				value:    "",
				register: false,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+" + tagEnabledName + "=package",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: false,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+" + tagEnabledName + "=package,register",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: true,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+" + tagEnabledName + "=package,register=true",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: true,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+" + tagEnabledName + "=package,register=false",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: false,
			},
		},
	}

	for i, tc := range testCases {
		r := extractEnabledTag(tc.comments)
		if r == nil && tc.expect != nil {
			t.Errorf("case[%d]: expected non-nil", i)
		}
		if r != nil && tc.expect == nil {
			t.Errorf("case[%d]: expected nil, got %v", i, *r)
		}
		if r != nil && *r != *tc.expect {
			t.Errorf("case[%d]: expected %v, got %v", i, *tc.expect, *r)
		}
	}
}

func Test_HasEqual(t *testing.T) {
	testCases := []struct {
		typ              *types.Type
		name             string
		hasEqual         bool
		pointerParameter bool
	}{
		{
			typ: &types.Type{
				Methods: map[string]*types.Type{},
			},
			name: "no methods",
		},
		{
			typ: &types.Type{
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Byte,
							},
						},
					},
				},
			},
			name: "Equal method with wrong result type",
		},
		{
			typ: &types.Type{
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Bool, types.Byte,
							},
						},
					},
				},
			},
			name: "Equal method with wrong number of results",
		},
		{
			typ: &types.Type{
				Name: types.Name{Package: "foo", Name: "Bar"},
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Bool,
							},
							Parameters: []*types.Type{
								{
									Name: types.Name{Package: "wrong", Name: "Bar"},
								},
							},
						},
					},
				},
			},
			name: "Equal method with correct result but wrong parameter type",
		},
		{
			typ: &types.Type{
				Name: types.Name{Package: "foo", Name: "Bar"},
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Bool,
							},
							Parameters: []*types.Type{
								{
									Name: types.Name{Package: "foo", Name: "Bar"},
								},
								{
									Name: types.Name{Package: "wrong", Name: "What"},
								},
							},
						},
					},
				},
			},
			name: "Equal method with correct result but wrong number of parameters",
		},
		{
			typ: &types.Type{
				Name: types.Name{Package: "foo", Name: "Bar"},
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Bool,
							},
							Parameters: []*types.Type{
								{
									Name: types.Name{Package: "foo", Name: "Bar"},
									Kind: types.Pointer,
								},
							},
						},
					},
				},
			},
			name:             "Equal method with correct result and pointer parameter type",
			hasEqual:         true,
			pointerParameter: true,
		},
		{
			typ: &types.Type{
				Name: types.Name{Package: "foo", Name: "Bar"},
				Methods: map[string]*types.Type{
					"Equal": {
						Signature: &types.Signature{
							Results: []*types.Type{
								types.Bool,
							},
							Parameters: []*types.Type{
								{
									Name: types.Name{Package: "foo", Name: "Bar"},
									Kind: types.Struct,
								},
							},
						},
					},
				},
			},
			name:     "Equal method with correct result and value parameter type",
			hasEqual: true,
		},
	}

	for _, tc := range testCases {
		hasEqual, pointer := HasEqual(tc.typ)
		if hasEqual != tc.hasEqual || pointer != tc.pointerParameter {
			t.Errorf("Expected hasEqual=%t and pointer=%t, got hasEqual=%t and pointer=%t", tc.hasEqual, tc.pointerParameter, hasEqual, pointer)
		}

	}
}
