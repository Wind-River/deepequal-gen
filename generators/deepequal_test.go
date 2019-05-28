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

func Test_isRootedUnder(t *testing.T) {
	testCases := []struct {
		path   string
		roots  []string
		expect bool
	}{
		{
			path:   "/foo/bar",
			roots:  nil,
			expect: false,
		},
		{
			path:   "/foo/bar",
			roots:  []string{},
			expect: false,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"/bad",
			},
			expect: false,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"/foo",
			},
			expect: true,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"/bad",
				"/foo",
			},
			expect: true,
		},
		{
			path: "/foo/bar/qux/zorb",
			roots: []string{
				"/foo/bar/qux",
			},
			expect: true,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"/foo/bar",
			},
			expect: true,
		},
		{
			path: "/foo/barn",
			roots: []string{
				"/foo/bar",
			},
			expect: false,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"/foo/barn",
			},
			expect: false,
		},
		{
			path: "/foo/bar",
			roots: []string{
				"",
			},
			expect: true,
		},
	}

	for i, tc := range testCases {
		r := isRootedUnder(tc.path, tc.roots)
		if r != tc.expect {
			t.Errorf("case[%d]: expected %t, got %t for %q in %q", i, tc.expect, r, tc.path, tc.roots)
		}
	}
}

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
				// No DeepCopyInto method.
				Methods: map[string]*types.Type{},
			},
			expect: false,
		},
		{
			typ: types.Type{
				Name: types.Name{Package: "pkgname", Name: "typename"},
				Kind: types.Builtin,
				Methods: map[string]*types.Type{
					// No DeepCopyInto method.
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
					"DeepCopyInto": {
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
					"DeepCopyInto": {
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
					"DeepCopyInto": {
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
					"DeepCopyInto": {
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
					"DeepCopyInto": {
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
							Results: []*types.Type{},
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
					"DeepCopyInto": {
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
							Results: []*types.Type{},
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
				"+k8s:deepcopy-gen",
			},
			expect: &enabledTagValue{
				value:    "",
				register: false,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+k8s:deepcopy-gen=package",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: false,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+k8s:deepcopy-gen=package,register",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: true,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+k8s:deepcopy-gen=package,register=true",
			},
			expect: &enabledTagValue{
				value:    "package",
				register: true,
			},
		},
		{
			comments: []string{
				"Human comment",
				"+k8s:deepcopy-gen=package,register=false",
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
