/*
SPDX-License-Identifier: Apache-2.0
Copyright 2016 The Kubernetes Authors.
Copyright 2019 Wind River Systems, Inc.
*/

package generators

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"k8s.io/gengo/args"
	"k8s.io/gengo/examples/set-gen/sets"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"k8s.io/klog"
)

// CustomArgs is used tby the go2idl framework to pass args specific to this
// generator.
type CustomArgs struct {
	BoundingDirs   []string // Only deal with types rooted under these dirs.
	GenPackagePath string   // Overwritten package path to be generated
}

// This is the comment tag that carries parameters for deep-copy generation.
const (
	tagEnabledName            = "deepequal-gen"
	tagIgnoreNilFieldsTagName = tagEnabledName + ":ignore-nil-fields"
	tagUnorderedArraysTagName = tagEnabledName + ":unordered-array"
)

// Known values for the comment tag.
const tagValuePackage = "package"

// enabledTagValue holds parameters from a tagName tag.
type enabledTagValue struct {
	value    string
	register bool
}

func extractEnabledTypeTag(t *types.Type) *enabledTagValue {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	return extractEnabledTag(comments)
}

func extractEnabledTag(comments []string) *enabledTagValue {
	tagVals := types.ExtractCommentTags("+", comments)[tagEnabledName]
	if tagVals == nil {
		// No match for the tag.
		return nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		klog.Fatalf("Found %d %s tags: %q", len(tagVals), tagEnabledName, tagVals)
	}

	// If we got here we are returning something.
	tag := &enabledTagValue{}

	// Get the primary value.
	parts := strings.Split(tagVals[0], ",")
	if len(parts) >= 1 {
		tag.value = parts[0]
	}

	// Parse extra arguments.
	parts = parts[1:]
	for i := range parts {
		kv := strings.SplitN(parts[i], "=", 2)
		k := kv[0]
		v := ""
		if len(kv) == 2 {
			v = kv[1]
		}
		switch k {
		case "register":
			if v != "false" {
				tag.register = true
			}
		default:
			klog.Fatalf("Unsupported %s param: %q", tagEnabledName, parts[i])
		}
	}
	return tag
}

func extractUnorderedArrayTypeTag(t *types.Type) *enabledTagValue {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	return extractUnorderedArrayTag(comments)
}

func extractUnorderedArrayTag(comments []string) *enabledTagValue {
	tagVals := types.ExtractCommentTags("+", comments)[tagUnorderedArraysTagName]
	if tagVals == nil {
		// No match for the tag.
		return nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		klog.Fatalf("Found %d %s tags: %q", len(tagVals), tagUnorderedArraysTagName, tagVals)
	}

	// If we got here we are returning something.
	tag := &enabledTagValue{}

	// Get the tag value.
	parts := strings.Split(tagVals[0], ",")
	if len(parts) > 1 {
		klog.Fatalf("Found %d %s tag values: %q", len(tagVals[0]), tagUnorderedArraysTagName, tagVals)
	}

	tag.value = parts[0]

	return tag
}

func extractIgnoreNilFieldsTypeTag(t *types.Type) *enabledTagValue {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	return extractIgnoreNilFieldsTag(comments)
}

func extractIgnoreNilFieldsTag(comments []string) *enabledTagValue {
	tagVals := types.ExtractCommentTags("+", comments)[tagIgnoreNilFieldsTagName]
	if tagVals == nil {
		// No match for the tag.
		return nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		klog.Fatalf("Found %d %s tags: %q", len(tagVals), tagIgnoreNilFieldsTagName, tagVals)
	}

	// If we got here we are returning something.
	tag := &enabledTagValue{}

	// Get the tag value.
	parts := strings.Split(tagVals[0], ",")
	if len(parts) > 1 {
		klog.Fatalf("Found %d %s tag values: %q", len(tagVals[0]), tagIgnoreNilFieldsTagName, tagVals)
	}

	tag.value = parts[0]

	return tag
}

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public": namer.NewPublicNamer(1),
		"raw":    namer.NewRawNamer("", nil),
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

func Packages(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		klog.Fatalf("Failed loading boilerplate: %v", err)
	}

	inputs := sets.NewString(context.Inputs...)
	packages := generator.Packages{}
	header := append([]byte(fmt.Sprintf("// +build !%s\n\n", arguments.GeneratedBuildTag)), boilerplate...)

	boundingDirs := make([]string, 0)
	if customArgs, ok := arguments.CustomArgs.(*CustomArgs); ok {
		if customArgs.BoundingDirs == nil {
			customArgs.BoundingDirs = context.Inputs
		}
		for i := range customArgs.BoundingDirs {
			// Strip any trailing slashes - they are not exactly "correct" but
			// this is friendlier.
			boundingDirs = append(boundingDirs, strings.TrimRight(customArgs.BoundingDirs[i], "/"))
		}
	}

	// Obtain override package path value
	genPackagePath := ""
	if customArgs, ok := arguments.CustomArgs.(*CustomArgs); ok {
		if len(customArgs.GenPackagePath) > 0 {
			genPackagePath = customArgs.GenPackagePath
		}
	}

	for i := range inputs {
		klog.V(5).Infof("Considering pkg %q", i)
		pkg := context.Universe[i]
		if pkg == nil {
			// If the input had no Go files, for example.
			continue
		}

		ptag := extractEnabledTag(pkg.Comments)
		ptagValue := ""
		ptagRegister := false
		if ptag != nil {
			ptagValue = ptag.value
			if ptagValue != tagValuePackage {
				klog.Fatalf("Package %v: unsupported %s value: %q", i, tagEnabledName, ptagValue)
			}
			ptagRegister = ptag.register
			klog.V(5).Infof("  tag.value: %q, tag.register: %t", ptagValue, ptagRegister)
		} else {
			klog.V(5).Infof("  no tag")
		}

		// If the pkg-scoped tag says to generate, we can skip scanning types.
		pkgNeedsGeneration := ptagValue == tagValuePackage
		if !pkgNeedsGeneration {
			// If the pkg-scoped tag did not exist, scan all types for one that
			// explicitly wants generation.
			for _, t := range pkg.Types {
				klog.V(5).Infof("  considering type %q", t.Name.String())
				ttag := extractEnabledTypeTag(t)
				if ttag != nil && ttag.value == "true" {
					klog.V(5).Infof("    tag=true")
					if !comparableType(t) {
						klog.Fatalf("Type %v requests deepEqual generation but is not copyable", t)
					}
					pkgNeedsGeneration = true
					break
				}
			}
		}

		if pkgNeedsGeneration {
			klog.V(3).Infof("Package %q needs generation", i)
			path := pkg.Path
			// if the source path is within a /vendor/ directory (for example,
			// k8s.io/kubernetes/vendor/k8s.io/apimachinery/pkg/apis/meta/v1), allow
			// generation to output to the proper relative path (under vendor).
			// Otherwise, the generator will create the file in the wrong location
			// in the output directory.
			// TODO: build a more fundamental concept in gengo for dealing with modifications
			// to vendored packages.
			if strings.HasPrefix(pkg.SourcePath, arguments.OutputBase) {
				expandedPath := strings.TrimPrefix(pkg.SourcePath, arguments.OutputBase)
				if strings.Contains(expandedPath, "/vendor/") {
					path = expandedPath
				}
			}
			// Set override package path if it is set.
			if len(genPackagePath) > 0 {
				path = genPackagePath
			}
			packages = append(packages,
				&generator.DefaultPackage{
					PackageName: strings.Split(filepath.Base(pkg.Path), ".")[0],
					PackagePath: path,
					HeaderText:  header,
					GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
						return []generator.Generator{
							NewGenDeepEqual(arguments.OutputFileBaseName, pkg.Path, boundingDirs, ptagValue == tagValuePackage, ptagRegister),
						}
					},
					FilterFunc: func(c *generator.Context, t *types.Type) bool {
						return t.Name.Package == pkg.Path
					},
				})
		}
	}
	return packages
}

// genDeepEqual produces a file with autogenerated deep-copy functions.
type genDeepEqual struct {
	generator.DefaultGen
	targetPackage string
	boundingDirs  []string
	allTypes      bool
	registerTypes bool
	imports       namer.ImportTracker
}

func NewGenDeepEqual(sanitizedName, targetPackage string, boundingDirs []string, allTypes, registerTypes bool) generator.Generator {
	return &genDeepEqual{
		DefaultGen: generator.DefaultGen{
			OptionalName: sanitizedName,
		},
		targetPackage: targetPackage,
		boundingDirs:  boundingDirs,
		allTypes:      allTypes,
		registerTypes: registerTypes,
		imports:       generator.NewImportTracker(),
	}
}

func (g *genDeepEqual) Namers(c *generator.Context) namer.NameSystems {
	// Have the raw namer for this file track what it imports.
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.targetPackage, g.imports),
	}
}

func (g *genDeepEqual) Filter(c *generator.Context, t *types.Type) bool {
	// Filter other types not being processed or not copyable within the package.
	enabled := g.allTypes
	if !enabled {
		ttag := extractEnabledTypeTag(t)
		if ttag != nil && ttag.value == "true" {
			enabled = true
		}
	}
	if !enabled {
		return false
	}
	if !comparableType(t) {
		klog.V(2).Infof("Type %v is not comparable", t)
		return false
	}
	klog.V(4).Infof("Type %v is copyable", t)
	return true
}

// deepEqualMethod returns the signature of a DeepEqual() method, nil or an error
// if the type is wrong. DeepEqual allows more efficient deep copy
// implementations to be defined by the type's author.  The correct signature
// for a type T is:
//    func (t T) DeepEqual(t *T)
// or:
//    func (t *T) DeepEqual(t *T)
func deepEqualMethod(t *types.Type) (*types.Signature, error) {
	f, found := t.Methods["DeepEqual"]
	if !found {
		return nil, nil
	}
	if len(f.Signature.Parameters) != 1 {
		return nil, fmt.Errorf("type %v: invalid DeepEqual signature, expected exactly one parameter", t)
	}
	if len(f.Signature.Results) != 1 || f.Signature.Results[0].Name.Name != "bool" {
		return nil, fmt.Errorf("type %v: invalid DeepEqual signature, expected bool result type", t)
	}

	ptrParam := f.Signature.Parameters[0].Kind == types.Pointer && f.Signature.Parameters[0].Elem.Name == t.Name

	if !ptrParam {
		return nil, fmt.Errorf("type %v: invalid DeepEqual signature, expected parameter of type *%s", t, t.Name.Name)
	}

	ptrRcvr := f.Signature.Receiver != nil && f.Signature.Receiver.Kind == types.Pointer && f.Signature.Receiver.Elem.Name == t.Name
	nonPtrRcvr := f.Signature.Receiver != nil && f.Signature.Receiver.Name == t.Name

	if !ptrRcvr && !nonPtrRcvr {
		// this should never happen
		return nil, fmt.Errorf("type %v: invalid DeepEqual signature, expected a receiver of type %s or *%s", t, t.Name.Name, t.Name.Name)
	}

	return f.Signature, nil
}

// deepEqualMethodOrDie returns the signature of a DeepEqualInto() method, nil or calls klog.Fatalf
// if the type is wrong.
func deepEqualMethodOrDie(t *types.Type) *types.Signature {
	ret, err := deepEqualMethod(t)
	if err != nil {
		klog.Fatal(err)
	}
	return ret
}

func comparableType(t *types.Type) bool {
	// If the type opts out of deepequal-generation, stop.
	ttag := extractEnabledTypeTag(t)
	if ttag != nil && ttag.value == "false" {
		return false
	}

	// Filter other private types.
	if namer.IsPrivateGoName(t.Name.Name) {
		return false
	}

	if t.Kind == types.Alias {
		// if the underlying built-in is not deepEqual-able, deepEqual is opt-in through definition of custom methods.
		// Note that aliases of builtins, maps, slices can have deepEqual methods.
		if deepEqualMethodOrDie(t) != nil {
			return true
		} else if t.Underlying.Kind == types.Pointer {
			return false
		} else if t.Underlying.Kind == types.Interface {
			return false
		} else {
			return t.Underlying.Kind != types.Builtin || comparableType(t.Underlying)
		}
	}

	if t.Kind != types.Struct {
		return false
	}

	return true
}

func underlyingType(t *types.Type) *types.Type {
	for t.Kind == types.Alias {
		t = t.Underlying
	}
	return t
}

func (g *genDeepEqual) isOtherPackage(pkg string) bool {
	if pkg == g.targetPackage {
		return false
	}
	if strings.HasSuffix(pkg, "\""+g.targetPackage+"\"") {
		return false
	}
	return true
}

func (g *genDeepEqual) Imports(c *generator.Context) (imports []string) {
	importLines := make([]string, 0)
	for _, singleImport := range g.imports.ImportLines() {
		if g.isOtherPackage(singleImport) {
			importLines = append(importLines, singleImport)
		}
	}
	return importLines
}

func argsFromType(ts ...*types.Type) generator.Args {
	a := generator.Args{
		"type": ts[0],
	}
	for i, t := range ts {
		a[fmt.Sprintf("type%d", i+1)] = t
	}
	return a
}

func (g *genDeepEqual) Init(c *generator.Context, w io.Writer) error {
	return nil
}

func (g *genDeepEqual) needsGeneration(t *types.Type) bool {
	tag := extractEnabledTypeTag(t)
	tv := ""
	if tag != nil {
		tv = tag.value
		if tv != "true" && tv != "false" {
			klog.Fatalf("Type %v: unsupported %s value: %q", t, tagEnabledName, tag.value)
		}
	}
	if g.allTypes && tv == "false" {
		// The whole package is being generated, but this type has opted other.
		klog.V(5).Infof("Not generating for type %v because type opted other", t)
		return false
	}
	if !g.allTypes && tv != "true" {
		// The whole package is NOT being generated, and this type has NOT opted in.
		klog.V(5).Infof("Not generating for type %v because type did not opt in", t)
		return false
	}
	return true
}

func (g *genDeepEqual) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	if !g.needsGeneration(t) {
		return nil
	}
	klog.V(5).Infof("Generating deepequal function for type %v", t)

	sw := generator.NewSnippetWriter(w, c, "$", "$")
	typeArgs := argsFromType(t)

	if deepEqualMethodOrDie(t) == nil {
		sw.Do("// DeepEqual is an autogenerated deepequal function, deeply comparing the \n", nil)
		sw.Do("// receiver with other. in must be non-nil.\n", nil)
		sw.Do("func (in *$.type|raw$) DeepEqual(other *$.type|raw$) bool {\n", typeArgs)
		g.generateFor(t, sw)
		sw.Do("\nreturn true\n", nil)
		sw.Do("}\n\n", nil)
	}

	// Create a fake entry for the type we just generated so that it gets
	// reused throughout the generated code.
	if t.Methods == nil {
		t.Methods = make(map[string]*types.Type)
	}

	t.Methods["DeepEqual"] = &types.Type{
		Kind: types.Func,
		Signature: &types.Signature{
			Receiver: &types.Type{
				Name: t.Name,
				Kind: types.Pointer,
				Elem: &types.Type{Name: t.Name},
			},
			Parameters: []*types.Type{{
				Kind: types.Pointer,
				Elem: &types.Type{Name: t.Name},
			}},
			Results: []*types.Type{types.Bool},
		},
	}
	return sw.Error()
}

// we use the system of shadowing 'in' and 'other' so that the same code is valid
// at any nesting level. This makes the autogenerator easy to understand, and
// the compiler shouldn't care.
func (g *genDeepEqual) generateFor(t *types.Type, sw *generator.SnippetWriter) {
	// derive inner types if t is an alias. We call the do* methods below with the alias type.
	// basic rule: generate according to inner type, but construct objects with the alias type.
	ut := underlyingType(t)

	var f func(*types.Type, *generator.SnippetWriter)
	switch ut.Kind {
	case types.Builtin:
		f = g.doBuiltin
	case types.Map:
		f = g.doMap
	case types.Slice:
		f = g.doSlice
	case types.Struct:
		f = g.doStruct
	case types.Pointer:
		f = g.doPointer
	case types.Interface:
		// interfaces are handled in-line in the other cases
		panic(fmt.Sprintf("Hit an interface type %v. This should never happen.", t))
	case types.Alias:
		// can never happen because we branch on the underlying type which is never an alias
		panic(fmt.Sprintf("Hit an alias type %v. This should never happen.", t))
	default:
		klog.Fatalf("Hit an unsupported type %v.", t)
	}

	if f != nil {
		f(t, sw)
	}
}

// doBuiltin generates code for a builtin or an alias to a builtin. The generated code is
// is the same for both cases, i.e. it's the code for the underlying type.
func (g *genDeepEqual) doBuiltin(t *types.Type, sw *generator.SnippetWriter) {
	sw.Do("if other == nil || in != *other {\n", nil)
	sw.Do("return false\n", nil)
	sw.Do("}\n", nil)
}

// doMap generates code for a map or an alias to a map. The generated code is
// is the same for both cases, i.e. it's the code for the underlying type.
func (g *genDeepEqual) doMap(t *types.Type, sw *generator.SnippetWriter) {
	ut := underlyingType(t)
	uet := underlyingType(ut.Elem)

	if deepEqualMethodOrDie(t) != nil {
		sw.Do("if other == nil || !in.DeepEqual(other) {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n", nil)
		return
	} else {
		sw.Do("if other == nil {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n\n", nil)
	}

	sw.Do("if len(*in) != len(*other) {\n", nil)
	sw.Do("return false\n", nil)
	sw.Do("} else {\n", nil)

	sw.Do("for key, inValue := range *in {\n", nil)
	sw.Do("if otherValue, present := (*other)[key]; !present {\n", nil)
	sw.Do("return false\n", nil)
	sw.Do("} else {\n", nil)
	if uet.IsPrimitive() {
		sw.Do("if inValue != otherValue {\n", nil)
	} else if uet.Kind == types.Pointer {
		if uet.Elem.IsPrimitive() {
			sw.Do("if ((inValue == nil) != (otherValue == nil) || ((inValue != nil) && (otherValue != nil) && (*inValue != *otherValue))) {\n", nil)
		} else {
			sw.Do("if !inValue.DeepEqual(otherValue) {\n", nil)
		}
	} else if ut.Elem.Kind != types.Alias && uet.Kind != types.Struct {
		// TODO(alegacy): for now we do not support generating an inline
		//  comparison for a complex structure.  The recommended approach is
		//  to define a type alias and to either manually define a DeepEqual
		//  method for it or to have one code generated.
		klog.Fatalf("Hit an unsupported type %v for %v, from %v", uet, ut, t)
	} else {
		sw.Do("if !inValue.DeepEqual(&otherValue) {\n", nil)
	}
	sw.Do("return false\n", nil)
	sw.Do("}\n", nil)
	sw.Do("}\n", nil)
	sw.Do("}\n", nil)
	sw.Do("}\n", nil)
}

// doSlice generates code for a slice or an alias to a slice. The generated code is
// is the same for both cases, i.e. it's the code for the underlying type.
func (g *genDeepEqual) doSlice(t *types.Type, sw *generator.SnippetWriter) {
	ut := underlyingType(t)
	uet := underlyingType(ut.Elem)

	if deepEqualMethodOrDie(t) != nil {
		sw.Do("if other == nil || !in.DeepEqual(other) {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n", nil)
		return
	} else {
		sw.Do("if other == nil {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n\n", nil)
	}

	unorderedArrayTag := extractUnorderedArrayTypeTag(t)

	sw.Do("if len(*in) != len(*other) {\n", nil)
	sw.Do("return false\n", nil)
	sw.Do("} else {\n", nil)
	if unorderedArrayTag != nil && unorderedArrayTag.value == "true" {
		sw.Do("for _, inElement := range *in {\n", nil)
		sw.Do("found := false\n", nil)
		sw.Do("for _, otherElement := range *other {\n", nil)
		if uet.IsPrimitive() {
			sw.Do("if inElement == otherElement {\n", nil)
		} else if uet.Kind == types.Pointer {
			if uet.Elem.IsPrimitive() {
				sw.Do("if ((inElement == nil) && (otherElement == nil) || ((inElement != nil) && (otherElement != nil) && (*inElement == *otherElement))) {\n", nil)
			} else {
				sw.Do("if inElement.DeepEqual(otherElement) {\n", nil)
			}
		} else if ut.Elem.Kind != types.Alias && uet.Kind != types.Struct {
			// TODO(alegacy): for now we do not support generating an inline
			//  comparison for a complex structure.  The recommended approach is
			//  to define a type alias and to either manually define a DeepEqual
			//  method for it or to have one code generated.
			klog.Fatalf("Hit an unsupported type %v for %v, from %v", uet, ut, t)
		} else {
			sw.Do("if inElement.DeepEqual(&otherElement) {\n", nil)
		}
		sw.Do("found = true\n", nil)
		sw.Do("break\n", nil)
		sw.Do("}\n", nil)
		sw.Do("}\n", nil)
		sw.Do("if !found {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n", nil)
		sw.Do("}\n", nil)
	} else {
		sw.Do("for i, inElement := range *in {\n", nil)
		if uet.IsPrimitive() {
			sw.Do("if inElement != (*other)[i] {\n", nil)
		} else if uet.Kind == types.Pointer {
			if uet.Elem.IsPrimitive() {
				sw.Do("if ((inElement == nil) && ((*other)[i] == nil) || ((inElement != nil) && ((*other)[i] != nil) && (*inElement != *(*other)[i]))) {\n", nil)
			} else {
				sw.Do("if !inElement.DeepEqual((*other)[i]) {\n", nil)
			}
		} else if ut.Elem.Kind != types.Alias && uet.Kind != types.Struct {
			// TODO(alegacy): for now we do not support generating an inline
			//  comparison for a complex structure.  The recommended approach is
			//  to define a type alias and to either manually define a DeepEqual
			//  method for it or to have one code generated.
			klog.Fatalf("Hit an unsupported type %v for %v, from %v", uet, ut, t)
		} else {
			sw.Do("if !inElement.DeepEqual(&(*other)[i]) {\n", nil)
		}
		sw.Do("return false\n", nil)
		sw.Do("}\n", nil)
		sw.Do("}\n", nil)
	}
	sw.Do("}\n", nil)
}

func HasEqual(t *types.Type) (hasEqual, pointerParameter bool) {
	method := t.Methods["Equal"]
	if method == nil {
		return
	}

	signature := method.Signature
	results := len(signature.Results) == 1 && signature.Results[0] == types.Bool
	parameters := len(signature.Parameters) == 1 && signature.Parameters[0].Name == t.Name
	hasEqual = results && parameters
	if hasEqual {
		pointerParameter = method.Signature.Parameters[0].Kind == types.Pointer
	}
	return
}

// IsAssignable returns whether the type is deep-assignable.  For example,
// slices and maps and pointers are shallow copies, but ints and strings are
// complete.
func IsComparable(t *types.Type) bool {
	if t.IsPrimitive() {
		return true
	}
	if t.Kind == types.Struct {
		for _, m := range t.Members {
			if !IsComparable(m.Type) {
				return false
			}
		}
		return true
	}
	return false
}

// doStruct generates code for a struct or an alias to a struct. The generated code is
// is the same for both cases, i.e. it's the code for the underlying type.
func (g *genDeepEqual) doStruct(t *types.Type, sw *generator.SnippetWriter) {
	ut := underlyingType(t)

	if deepEqualMethodOrDie(t) != nil {
		sw.Do("if other == nil || !in.DeepEqual(other) {\n", t)
		sw.Do("return false\n", nil)
		sw.Do("}\n", nil)
		return
	} else {
		sw.Do("if other == nil {\n", nil)
		sw.Do("return false\n", nil)
		sw.Do("}\n\n", nil)
	}

	ignoreNilFieldsTag := extractIgnoreNilFieldsTypeTag(ut)

	for _, m := range ut.Members {
		ft := m.Type
		uft := underlyingType(ft)

		typeArgs := generator.Args{
			"type": ft,
			"kind": ft.Kind,
			"name": m.Name,
		}

		switch {
		case uft.Kind == types.Builtin:
			sw.Do("if in.$.name$ != other.$.name$ {\n", typeArgs)
			sw.Do("return false\n", nil)
			sw.Do("}\n", nil)

		case uft.Kind == types.Pointer:
			ufet := underlyingType(uft.Elem)
			if ignoreNilFieldsTag != nil && ignoreNilFieldsTag.value == "true" {
				// The is some optional attribute that should not be considered
				// when it is nil.
				sw.Do("if in.$.name$ != nil {\n", typeArgs)
			}
			sw.Do("if (in.$.name$ == nil) != (other.$.name$ == nil) {\n", typeArgs)
			sw.Do("return false\n", nil)
			sw.Do("} else if in.$.name$ != nil {\n", typeArgs)
			if ufet.IsPrimitive() {
				sw.Do("if *in.$.name$ != *other.$.name$ {\n", typeArgs)
			} else {
				sw.Do("if !in.$.name$.DeepEqual(other.$.name$) {\n", typeArgs)
			}
			sw.Do("return false\n", nil)
			sw.Do("}\n", nil)
			sw.Do("}\n", nil)
			if ignoreNilFieldsTag != nil && ignoreNilFieldsTag.value == "true" {
				sw.Do("}\n", nil)
			}
			sw.Do("\n", nil)

		case uft.Kind == types.Slice, uft.Kind == types.Map:
			sw.Do("if ((in.$.name$ != nil) && (other.$.name$ != nil)) ||", typeArgs)
			sw.Do("((in.$.name$ == nil) != (other.$.name$ == nil)) {\n", typeArgs)
			sw.Do("in, other := &in.$.name$, &other.$.name$\n", typeArgs)
			g.generateFor(ft, sw)
			sw.Do("}\n\n", nil)

		case uft.Kind == types.Struct:
			if IsComparable(uft) {
				sw.Do("if in.$.name$ != other.$.name$ {\n", typeArgs)
			} else if hasEqual, pointer := HasEqual(uft); hasEqual {
				if pointer {
					sw.Do("if !in.$.name$.Equal(&other.$.name$) {\n", typeArgs)
				} else {
					sw.Do("if !in.$.name$.Equal(other.$.name$) {\n", typeArgs)
				}
			} else {
				sw.Do("if !in.$.name$.DeepEqual(&other.$.name$) {\n", typeArgs)
			}
			sw.Do("return false\n", nil)
			sw.Do("}\n\n", nil)

		case uft.Kind == types.Interface:
			fallthrough
		default:
			klog.Fatalf("Hit an unsupported type %v for %v, from %v", uft, ft, t)
		}
	}
}

// doPointer generates code for a pointer or an alias to a pointer. The generated code is
// is the same for both cases, i.e. it's the code for the underlying type.
func (g *genDeepEqual) doPointer(t *types.Type, sw *generator.SnippetWriter) {
	// TODO(alegacy)
	ut := underlyingType(t)
	klog.Fatalf("Hit an unsupported type %v from %v", ut, t)
}
