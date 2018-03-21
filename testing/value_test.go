// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

type fakeRandSource struct {
	v int64
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
func (fr *fakeRandSource) Seed(seed int64) {
	fr.v = seed
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (fr *fakeRandSource) Int63() int64 {
	return fr.v
}

func NewFakeRand(seed int64) *rand.Rand {
	fr := rand.New(&fakeRandSource{})
	fr.Seed(seed)
	return fr
}

type mapKey struct {
	Name string
	Ago  int16
}

type generatedType struct {
	A int64
	B int64
}

func (gt generatedType) Generate(r *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(generatedType{A: randInt64(r), B: randInt64(r)})
}

type generatedTypePtr struct {
	C int64
	D int64
}

func (gt *generatedTypePtr) Generate(r *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(&generatedTypePtr{C: randInt64(r), D: randInt64(r)})
}

type filledSmallStruct struct {
	internal string

	B bool
	I int
	U uint
}

type leaf struct {
	Value int
	Left  *leaf
	Right *leaf
}

type filledStruct struct {
	internal1 string
	internal2 bool

	B     bool
	I     int
	I8    int8
	I16   int16
	I32   int32
	I64   int64
	UI    uint
	UI8   uint8
	UI16  uint16
	UI32  uint32
	UI64  uint64
	UIPtr uintptr
	F32   float32
	F64   float64
	C64   complex64
	C128  complex128

	Bp     *bool
	Ip     *int
	I8p    *int8
	I16p   *int16
	I32p   *int32
	I64p   *int64
	UIp    *uint
	UI8p   *uint8
	UI16p  *uint16
	UI32p  *uint32
	UI64p  *uint64
	UIPtrp *uintptr
	F32p   *float32
	F64p   *float64
	C64p   *complex64
	C128p  *complex128

	Str string

	GT     generatedType
	GTp    *generatedType
	GTPtr  generatedTypePtr
	GTPtrP *generatedTypePtr

	AInt       [10]int
	AStr       [20]string
	AStruct    [30]filledSmallStruct
	MIntStr    map[int]string
	MStrStr    map[string]string
	MStructInt map[mapKey]uint64
	SByte      []byte
	SStr       []string
	SStruct    []*filledSmallStruct
	Struct     filledSmallStruct

	SGTp []*generatedType

	Leaf leaf
}

func traverseValue(v reflect.Value, name string, process func(v reflect.Value, name string) bool) bool {
	v = reflect.Indirect(v)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if !traverseValue(v.Index(i), fmt.Sprintf("%s[%d]", name, i), process) {
				return false
			}
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			if !traverseValue(v.MapIndex(k), fmt.Sprintf("%s[%v]", name, k), process) {
				return false
			}
		}
	case reflect.Struct:
		for i, n := 0, v.NumField(); i < n; i++ {
			if fieldValue := v.Field(i); fieldValue.CanSet() {
				if !traverseValue(fieldValue, fmt.Sprintf("%s.%s", name, v.Type().Field(i).Name), process) {
					return false
				}
			}
		}
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
	default:
		return process(v, name)
	}
	return true
}

func TestValue(t *testing.T) {

	config := NewValueConfig()
	config.Depth = 10
	config.Size = 10000
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 5000; i++ {
		//seed := rand.Int63()
		//seed := int64(1895008806655175727) // Ok
		//seed := int64(9221914701813994119) // time out
		//seed := int64(9221562872235418891) // time out
		//fmt.Printf("Seed: %d; i: %d\n", seed, i)
		//fakeRand := NewFakeRand(seed)
		fakeRand := NewFakeRand(rand.Int63())
		samples := map[reflect.Kind]*struct {
			t reflect.Type
			v reflect.Value
		}{
			reflect.Bool:       {reflect.TypeOf(bool(false)), reflect.ValueOf(fakeRand.Int()&1 == 0)},
			reflect.Int:        {reflect.TypeOf(int(0)), reflect.ValueOf(int(randInt64(fakeRand)))},
			reflect.Int8:       {reflect.TypeOf(int8(0)), reflect.ValueOf(int8(randInt64(fakeRand)))},
			reflect.Int16:      {reflect.TypeOf(int16(0)), reflect.ValueOf(int16(randInt64(fakeRand)))},
			reflect.Int32:      {reflect.TypeOf(int32(0)), reflect.ValueOf(int32(randInt64(fakeRand)))},
			reflect.Int64:      {reflect.TypeOf(int64(0)), reflect.ValueOf(randInt64(fakeRand))},
			reflect.Uint:       {reflect.TypeOf(uint(0)), reflect.ValueOf(uint(randInt64(fakeRand)))},
			reflect.Uint8:      {reflect.TypeOf(uint8(0)), reflect.ValueOf(uint8(randInt64(fakeRand)))},
			reflect.Uint16:     {reflect.TypeOf(uint16(0)), reflect.ValueOf(uint16(randInt64(fakeRand)))},
			reflect.Uint32:     {reflect.TypeOf(uint32(0)), reflect.ValueOf(uint32(randInt64(fakeRand)))},
			reflect.Uint64:     {reflect.TypeOf(uint64(0)), reflect.ValueOf(uint64(randInt64(fakeRand)))},
			reflect.Uintptr:    {reflect.TypeOf(uintptr(0)), reflect.ValueOf(uintptr(randInt64(fakeRand)))},
			reflect.Float32:    {reflect.TypeOf(float32(0)), reflect.ValueOf(randFloat32(fakeRand))},
			reflect.Float64:    {reflect.TypeOf(float64(0)), reflect.ValueOf(randFloat64(fakeRand))},
			reflect.Complex64:  {reflect.TypeOf(complex64(0)), reflect.ValueOf(complex(randFloat32(fakeRand), randFloat32(fakeRand)))},
			reflect.Complex128: {reflect.TypeOf(complex128(0)), reflect.ValueOf(complex(randFloat64(fakeRand), randFloat64(fakeRand)))},
			reflect.String:     {reflect.TypeOf(string("")), reflect.ValueOf(randString(fakeRand, config.MaxStringLength))},
		}

		// Test of creating simple types
		for _, sample := range samples {
			if v, ok := Value(sample.t, fakeRand, config); ok {
				if !reflect.DeepEqual(sample.v.Interface(), v.Interface()) {
					t.Errorf("Value must be equal to %[1]T(%#[1]v), but it has a value %[2]T(%#[2]v)", sample.v.Interface(), v.Interface())
					return
				}
			} else {
				t.Errorf("Unable to create arbitrary value of type %s. Value returned false.", sample.t)
				return
			}
		}

		// Test of creating structure
		checkValue := func(v reflect.Value, name string) bool {
			if sample, ok := samples[v.Kind()]; ok {
				b := reflect.DeepEqual(sample.v.Interface(), v.Interface())
				if !b {
					t.Errorf("Value of \"%s\" must be equal to %[2]T(%#[2]v), but it has a value %[3]T(%#[3]v)", name, sample.v.Interface(), v.Interface())
				}
				return b
			}
			return true
		}

		if v, ok := Value(reflect.TypeOf(filledStruct{}), fakeRand, config); ok {
			if !traverseValue(v, "", checkValue) {
				return
			}
		} else {
			t.Errorf("Unable to create arbitrary value. Value returned false.")
			return
		}

		if v, ok := Value(timeType, fakeRand, config); ok {
			timeValue := reflect.ValueOf(randTime(fakeRand, config.MinTime, config.MaxTime))
			b := reflect.DeepEqual(v.Interface(), timeValue.Interface())
			if !b {
				t.Errorf("Value must be equal to %[1]T(%#[1]v), but it has a value %[2]T(%#[2]v)", timeValue.Interface(), v.Interface())
				return
			}
		} else {
			t.Errorf("Unable to create arbitrary value. Value returned false.")
			return
		}

	}

}
