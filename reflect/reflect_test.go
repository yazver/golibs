package reflect

import (
	"errors"
	"net"
	"reflect"
	"testing"
	"time"
)

func checkEqual(t *testing.T, actual, expected interface{}) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Not equal. Actual: %v; expected: %v", actual, expected)
	}
}

func Test_AssignString(t *testing.T) {
	var i int
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var ui uint
	var ui8 uint8
	var ui16 uint16
	var ui32 uint32
	var ui64 uint64
	var f32 float32
	var f64 float64
	var b bool
	var s string
	var ti time.Time
	var d time.Duration
	var ip net.IP

	type args struct {
		dst      interface{}
		src      string
		expected interface{}
	}

	tests := []struct {
		args    args
		wantErr bool
	}{
		{args{&i, "8546778", int(8546778)}, false},
		{args{&i8, "16", int8(16)}, false},
		{args{&i16, "-30000", int16(-30000)}, false},
		{args{&i32, "0x7fffffff", int32(0x7fffffff)}, false},
		{args{&i64, "-0xffffffff", int64(-0xffffffff)}, false},

		{args{&i, "100.1", nil}, true},
		{args{&i8, "255", nil}, true},
		{args{&i16, "0xffff", nil}, true},
		{args{&i32, "0xffffffff", nil}, true},
		{args{&i64, "true", nil}, true},

		{args{&ui, "8546778", uint(8546778)}, false},
		{args{&ui8, "16", uint8(16)}, false},
		{args{&ui16, "64000", uint16(64000)}, false},
		{args{&ui32, "0xffffffff", uint32(0xffffffff)}, false},
		{args{&ui64, "0xffffffffffffffff", uint64(0xffffffffffffffff)}, false},

		{args{&ui, "100.1", nil}, true},
		{args{&ui8, "-255", nil}, true},
		{args{&ui16, "0xffffffff", nil}, true},
		{args{&ui32, "0x8ffffffff", nil}, true},
		{args{&ui64, "fish", nil}, true},

		{args{&f32, "10.1", float32(10.1)}, false},
		{args{&f64, "-5.12345678e42", float64(-5.12345678e42)}, false},

		{args{&f32, "rabbit", nil}, true},
		{args{&f64, "5.1234.5678", nil}, true},

		{args{&b, "true", bool(true)}, false},
		{args{&b, "0", bool(false)}, false},

		{args{&b, "FaLsE", nil}, true},
		{args{&b, "10", nil}, true},

		{args{&s, "I DOWN THE RABBIT HOLE", nil}, false},

		{args{&ti, "1832-01-27T01:02:03+05:00", nil}, false},
		{args{&ti, "1832-01-27T01-02-03", nil}, true},

		{args{&d, "22h49m22s0ms", time.Hour*22 + time.Minute*49 + time.Second*22}, false},
		{args{&d, "25r", nil}, true},

		{args{&ip, "127.0.0.1", net.IPv4(127, 0, 0, 1)}, false},
		{args{&ip, "2001:db8::", net.ParseIP("2001:db8::")}, false},
		{args{&ip, "1000.0.0.1", nil}, true},

		{args{s, "unassignable", nil}, true},
	}

	for _, tt := range tests {
		dst := reflect.Indirect(reflect.ValueOf(tt.args.dst))
		if err := AssignString(tt.args.dst, tt.args.src); (err != nil) != tt.wantErr {
			t.Errorf("AssignString(%v, \"%v\"), error = %v, wantErr %v",
				dst.Type().Name(), tt.args.src, err, tt.wantErr)
		} else if err == nil && tt.args.expected != nil {
			checkEqual(t, dst.Interface(), tt.args.expected)
		}
	}
}

func Test_AssignValue(t *testing.T) {
	var i int
	var i8 int8
	var i16 int16
	var i32 int32
	var i64 int64
	var ui uint
	var ui8 uint8
	var ui16 uint16
	var ui32 uint32
	var ui64 uint64
	var f32 float32
	var f64 float64
	var b bool
	var s string
	var ti time.Time
	var d time.Duration
	var ip net.IP

	b2 := true

	type args struct {
		dst      interface{}
		src      interface{}
		expected interface{}
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{args{&i, "8546778", int(8546778)}, false},
		{args{&i8, "16", int8(16)}, false},
		{args{&i16, "-30000", int16(-30000)}, false},
		{args{&i32, "0x7fffffff", int32(0x7fffffff)}, false},
		{args{&i64, "-0xffffffff", int64(-0xffffffff)}, false},
		{args{&i, int(-100), int(-100)}, false},
		{args{&i8, &i, int8(-100)}, false}, // from previous
		{args{&i16, int8(9), int16(9)}, false},
		{args{&i32, int64(1000000000), int32(1000000000)}, false},
		{args{&i64, &i32, int64(1000000000)}, false}, // from previous

		{args{&i, "100.1", nil}, true},
		{args{&i8, "255", nil}, true},
		{args{&i16, "0xffff", nil}, true},
		{args{&i32, "0xffffffff", nil}, true},
		{args{&i64, "true", nil}, true},
		{args{&i, 1 - 0.707i, nil}, true},
		{args{&i8, ip, nil}, true},
		{args{&i16, &ti, nil}, true},
		{args{&i32, struct{}{}, nil}, true},
		{args{&i64, &b, nil}, true},

		{args{&ui, "8546778", uint(8546778)}, false},
		{args{&ui8, "16", uint8(16)}, false},
		{args{&ui16, "64000", uint16(64000)}, false},
		{args{&ui32, "0xffffffff", uint32(0xffffffff)}, false},
		{args{&ui64, "0xffffffffffffffff", uint64(0xffffffffffffffff)}, false},
		{args{&ui, uint(100), uint(100)}, false},
		{args{&ui8, uint16(90), uint8(90)}, false},
		{args{&ui16, &ui8, uint16(90)}, false}, // from previous
		{args{&ui32, uint16(64000), uint32(64000)}, false},
		{args{&ui64, &ui32, uint64(64000)}, false}, // from previous

		{args{&ui, "100.1", nil}, true},
		{args{&ui8, "-255", nil}, true},
		{args{&ui16, "0xffffffff", nil}, true},
		{args{&ui32, "0x8ffffffff", nil}, true},
		{args{&ui64, "fish", nil}, true},
		{args{&ui, struct{}{}, nil}, true},
		{args{&ui8, "1.5cow", nil}, true},
		{args{&ui16, true, nil}, true},
		{args{&ui32, &ti, nil}, true},
		{args{&ui64, complex64(1.1), nil}, true},

		{args{&f32, "10.1", float32(10.1)}, false},
		{args{&f64, "-5.12345678e42", float64(-5.12345678e42)}, false},
		{args{&f32, int(100), float32(100)}, false},
		{args{&f64, &f32, nil}, false},
		{args{&f32, "rabbit", nil}, true},
		{args{&f64, "5.1234.5678", nil}, true},
		{args{&f32, true, nil}, true},
		{args{&f64, &ip, nil}, true},

		{args{&b, "true", true}, false},
		{args{&b, "0", false}, false},
		{args{&b, false, false}, false},
		{args{&b, &b2, b2}, false},

		{args{&b, "FaLsE", nil}, true},
		{args{&b, "10", nil}, true},
		{args{&b, uint(20), nil}, true},
		{args{&b, &f64, nil}, true},

		{args{&s, "I DOWN THE RABBIT HOLE", nil}, false},
		{args{&s, &i8, nil}, false},
		{args{&s, struct{}{}, nil}, true},
		{args{&s, "", nil}, false},
		{args{s, "I DOWN THE RABBIT HOLE", nil}, true},

		{args{&ti, "1832-01-27T01:02:03+05:00", time.Date(1832, 1, 27, 1, 2, 3, 0, time.FixedZone("", 5*60*60))}, false},
		{args{&ti, time.Now(), nil}, false},
		{args{&ti, "1832-01-27T01-02-03", nil}, true},
		{args{&ti, &b, nil}, true},

		{args{&d, "22h49m22s0ms", time.Hour*22 + time.Minute*49 + time.Second*22}, false},
		{args{&d, time.Hour, nil}, false},
		{args{&d, "25r", nil}, true},
		{args{&d, &ti, nil}, true},

		{args{&ip, "127.0.0.1", net.IPv4(127, 0, 0, 1)}, false},
		{args{&ip, "2001:db8::", net.ParseIP("2001:db8::")}, false},
		{args{&ip, net.IPv4(4, 31, 198, 44), net.IPv4(4, 31, 198, 44)}, false},
		{args{&ip, "1000.0.0.1", nil}, true},
		{args{&ip, d, nil}, true},

		{args{s, "unassignable", nil}, true},
		{args{new(struct{}), "hi", nil}, true},
	}
	for _, tt := range tests {
		dst := reflect.Indirect(reflect.ValueOf(tt.args.dst))
		if err := Assign(tt.args.dst, tt.args.src); (err != nil) != tt.wantErr {
			t.Errorf("Assign(%v, %v), error = %v, wantErr %v",
				dst.Type().Name(), reflect.TypeOf(tt.args.src).Name(), err, tt.wantErr)
		} else if err == nil && tt.args.expected != nil {
			checkEqual(t, dst.Interface(), tt.args.expected)
		}
	}
}

func Test_Traverse(t *testing.T) {
	type mini struct {
		I int64
	}
	s := struct {
		String string
		Int    int
		Map    map[int]bool
		Slice  []string
		M      mini
	}{"str", 2, map[int]bool{1: true, 2: false}, []string{"one", "two"}, mini{10000000000}}

	type args struct {
		expected  interface{}
		processed bool
	}
	ma := map[string]*args{
		"String":   &args{"str", false},
		"Int":      &args{2, false},
		"Map[1]":   &args{true, false},
		"Map[2]":   &args{false, false},
		"Slice[0]": &args{"one", false},
		"Slice[1]": &args{"two", false},
		"M.I":      &args{int64(10000000000), false},
	}

	process := func(value reflect.Value, path string, level uint, field *reflect.StructField) error {
		if a, ok := ma[path]; ok {
			if !reflect.DeepEqual(reflect.Indirect(value).Interface(), a.expected) {
				t.Errorf("Not equal. Actual: %v; expected: %v", reflect.Indirect(value).Interface(), a.expected)
			}
			a.processed = true
		}
		return nil
	}
	if err := Traverse(s, process); err != nil {
		t.Error(err)
	}
	for n, v := range ma {
		if !v.processed {
			t.Errorf("Value \"%v\" is not processed", n)
		}
	}

	// Check returning error
	st := struct {
		Slice []map[int]mini
	}{[]map[int]mini{map[int]mini{1: mini{1}}}}

	process = func(value reflect.Value, path string, level uint, field *reflect.StructField) error {
		if field != nil && field.Name == "I" {
			return errors.New("")
		}
		return nil
	}
	if err := Traverse(st, process); err == nil {
		t.Error("Must return error")
	}

}

func Test_TraverseFields(t *testing.T) {
	type mini struct {
		I int64
	}
	s := struct {
		String string
		Int    int
		Map    map[int]bool
		Slice  []string
		M      mini
	}{"str", 2, map[int]bool{1: true, 2: false}, []string{"one", "two"}, mini{10000000000}}

	type args struct {
		expected  interface{}
		processed bool
	}
	ma := map[string]*args{
		"String": &args{"str", false},
		"Int":    &args{2, false},
		"M.I":    &args{int64(10000000000), false},
	}

	process := func(value reflect.Value, path string, level uint, field *reflect.StructField) error {
		if field == nil {
			t.Error("It's not field: " + path)
		}
		if a, ok := ma[path]; ok {
			if !reflect.DeepEqual(reflect.Indirect(value).Interface(), a.expected) {
				t.Errorf("Not equal. Actual: %v; expected: %v", reflect.Indirect(value).Interface(), a.expected)
			}
			a.processed = true
		}
		return nil
	}
	if err := TraverseFields(s, process); err != nil {
		t.Error(err)
	}
	for n, v := range ma {
		if !v.processed {
			t.Errorf("Value \"%v\" is not processed", n)
		}
	}
}

func Test_Clear(t *testing.T) {
	i := 10
	pi := new(int)
	s := "Mu"
	Clear(&i)
	Clear(&pi)
	Clear(&s)
	if i != 0 {
		t.Error("\"i\" mast be equal 0")
	}
	if pi != nil {
		t.Error("\"pi\" mast be equal nil")
	}
	if s != "" {
		t.Error("\"s\" mast be equal \"\"")
	}
}
