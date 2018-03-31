package testing

import (
	"math"
	"math/rand"
	"reflect"
	"testing/quick"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})
 
// randFloat32 generates a random float taking the full range of a float32.
func randFloat32(rand *rand.Rand) float32 {
	f := rand.Float64() * math.MaxFloat32
	if rand.Int()&1 == 1 {
		f = -f
	}
	return float32(f)
}

// randFloat64 generates a random float taking the full range of a float64.
func randFloat64(rand *rand.Rand) float64 {
	f := rand.Float64() * math.MaxFloat64
	if rand.Int()&1 == 1 {
		f = -f
	}
	return f
}

// randInt64 returns a random integer taking half the range of an int64.
func randInt64(rand *rand.Rand) int64 { return rand.Int63() - 1<<62 }

// randString returns a random string.
func randString(rand *rand.Rand, maxSize int) string {
	numChars := rand.Int() % (maxSize + 1)
	//fmt.Printf("maxSize: %d; numChars: %d \n", maxSize, numChars)
	codePoints := make([]rune, numChars)
	for i := 0; i < numChars; i++ {
		codePoints[i] = rune(rand.Int()%0x10ffff + 1)
	}
	//fmt.Println("Exit randString")
	return string(codePoints)
}

// randTime returns a random Time.
func randTime(rand *rand.Rand, min, max time.Time) time.Time {
	if max.Before(min) {
		return min
	}
	if max.IsZero() {
		max = time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	minSec := min.Unix()
	sec := rand.Int63()%(max.Unix()-minSec) + minSec
	return time.Unix(sec, rand.Int63()%1000000000)
}

type GeneratorsList map[reflect.Type]func(t reflect.Type, rand *rand.Rand) (value reflect.Value, ok bool)

type ValueConfig struct {
	// The maximum depth of dive.
	// 0 - unlimited
	Depth int
	// The maximum number of created values.
	// 0 - unlimited
	Size int
	// Minimum length of arrays, slices and maps
	// default is 0
	MinLength int
	// Maximum length of arrays, slices and maps
	// default is 100
	MaxLength int
	// default is 100
	MaxStringLength int
	// The probability of the pointer to be a nil
	// 0 - never nil
	// 100 - always nil
	// default is 10
	NilProbability int
	// If is not defined then random Time will in 0001-01-01 00:00:00 +0000 UTC - 2100-01-01 00:00:00 +0000 UTC
	MinTime, MaxTime time.Time
	Generators       GeneratorsList
}

func (vc *ValueConfig) mustBeNil(rand *rand.Rand, depth int, size int) bool {
	if depth <= 0 || size <= 0 {
		return true
	} else if rand.Intn(100) < vc.NilProbability {
		return true
	}
	return false
}

func (vc *ValueConfig) length(rand *rand.Rand, depth int, size int) int {
	if depth > 0 && size > 0 {
		if vc.MinLength > size {
			return vc.MinLength
		}
		if l := vc.MinLength + rand.Intn(vc.MaxLength-vc.MinLength+1); (l > size) && (vc.Size != 0) {
			return size
		} else {
			return l
		}
	}
	return 0
}

func (vc *ValueConfig) generate(t reflect.Type, rand *rand.Rand) (value reflect.Value, ok bool) {
	if generator, ok := vc.Generators[t]; ok {
		return generator(t, rand)
	} else if generator, ok = vc.Generators[nil]; ok {
		return generator(t, rand)
	}
	return reflect.Value{}, false
}

func NewValueConfig() *ValueConfig {
	return &ValueConfig{
		Depth:           0,
		Size:            0,
		MinLength:       0,
		MaxLength:       100,
		MaxStringLength: 100,
		NilProbability:  10,
		MinTime:         time.Time{},
		MaxTime:         time.Time{},
		Generators:      GeneratorsList{},
	}
}

var defaultValueConfig = NewValueConfig()

// Value returns an arbitrary value of the given type.
// If the type implements the Generator interface, that will be used.
// Note: To create arbitrary values for structs, all the fields must be exported.
func Value(t reflect.Type, rand *rand.Rand, config *ValueConfig) (value reflect.Value, ok bool) {
	if config == nil {
		config = defaultValueConfig
	}

	const maxInt = int(^uint(0) >> 1)
	size := config.Size
	if size == 0 {
		size = maxInt
	}
	depth := config.Depth
	if depth == 0 {
		depth = maxInt
	}

	return generateValue(t, rand, config, depth, size)
}

func isCompositeType(t reflect.Type) bool {
	if t == timeType {
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.Struct:
		return true
	default:
		return false
	}
}

// generateValue returns an arbitrary value of the given type.
func generateValue(t reflect.Type, rand *rand.Rand, config *ValueConfig, depth int, size int) (value reflect.Value, ok bool) {
	size--
	depth--

	if value, ok = config.generate(t, rand); ok {
		return value, ok
	}

	if !(t.Kind() == reflect.Ptr && t.Elem().Implements(reflect.TypeOf((*quick.Generator)(nil)).Elem())) {
		if m, ok := reflect.Zero(t).Interface().(quick.Generator); ok {
			return m.Generate(rand, config.MaxLength), true
		}
	}

	if t == timeType {
		return reflect.ValueOf(randTime(rand, config.MinTime, config.MaxTime)), true
	}

	v := reflect.New(t).Elem()
	switch concrete := t; concrete.Kind() {
	case reflect.Bool:
		v.SetBool(rand.Int()&1 == 0)
	case reflect.Float32:
		v.SetFloat(float64(randFloat32(rand)))
	case reflect.Float64:
		v.SetFloat(randFloat64(rand))
	case reflect.Complex64:
		v.SetComplex(complex(float64(randFloat32(rand)), float64(randFloat32(rand))))
	case reflect.Complex128:
		v.SetComplex(complex(randFloat64(rand), randFloat64(rand)))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v.SetInt(randInt64(rand))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Uintptr:
		v.SetUint(uint64(randInt64(rand)))
	case reflect.Map:
		numElems := config.length(rand, depth, size)
		if numElems > 0 {
			sizeLeft := size / numElems
			if sizeLeft == 0 {
				sizeLeft = 1
			}
			v.Set(reflect.MakeMap(concrete))
			for i := 0; i < numElems; i++ {
				key, ok1 := generateValue(concrete.Key(), rand, config, depth, sizeLeft)
				value, ok2 := generateValue(concrete.Elem(), rand, config, depth, sizeLeft)
				if !ok1 || !ok2 {
					return reflect.Value{}, false
				}
				v.SetMapIndex(key, value)
			}
		}
	case reflect.Ptr:
		if config.mustBeNil(rand, depth, size) {
			v.Set(reflect.Zero(concrete)) // Generate nil pointer.
		} else {
			//*counter--
			elem, ok := generateValue(concrete.Elem(), rand, config, depth, size-1)
			if !ok {
				return reflect.Value{}, false
			}
			v.Set(reflect.New(concrete.Elem()))
			v.Elem().Set(elem)
		}
	case reflect.Slice:
		numElems := config.length(rand, depth, size)
		if numElems > 0 {
			sizeLeft := size / numElems
			if sizeLeft == 0 {
				sizeLeft = 1
			}
			v.Set(reflect.MakeSlice(concrete, numElems, numElems))
			for i := 0; i < numElems; i++ {
				elem, ok := generateValue(concrete.Elem(), rand, config, depth, sizeLeft)
				if !ok {
					return reflect.Value{}, false
				}
				v.Index(i).Set(elem)
			}
		}
	case reflect.Array:
		n := v.Len()
		if n > 0 {
			sizeLeft := size / n
			if sizeLeft == 0 {
				sizeLeft = 1
			}
			for i := 0; i < n; i++ {
				elem, ok := generateValue(concrete.Elem(), rand, config, depth, sizeLeft)
				if !ok {
					return reflect.Value{}, false
				}
				v.Index(i).Set(elem)
			}
		}
	case reflect.String:
		v.SetString(randString(rand, config.MaxStringLength))
	case reflect.Struct:
		n := v.NumField()
		if n > 0 {
			numCompositeTypes := 0
			numNonCompositeTypes := 0
			for i := 0; i < n; i++ {
				if fieldValue := v.Field(i); fieldValue.CanSet() {
					if isCompositeType(concrete.Field(i).Type) {
						numCompositeTypes++
					} else {
						numNonCompositeTypes++
					}
				}
			}

			sizeLeft := size - numNonCompositeTypes
			if sizeLeft > 0 {
				if numCompositeTypes > 0 {
					sizeLeft = sizeLeft / numCompositeTypes
					if sizeLeft == 0 {
						sizeLeft = 1
					}
				}
			} else {
				sizeLeft = 1
			}
			for i := 0; i < n; i++ {
				if fieldValue := v.Field(i); fieldValue.CanSet() {
					t := concrete.Field(i).Type
					fieldSize := 2
					if isCompositeType(t) {
						fieldSize = sizeLeft
						//fmt.Println(concrete.Field(i).Name, "; sizeLeft:", fieldSize)
					}
					elem, ok := generateValue(t, rand, config, depth, fieldSize)
					if !ok {
						return reflect.Value{}, false
					}
					fieldValue.Set(elem)
				}
			}
		}
	default:
		return reflect.Value{}, false
	}

	return v, true
}
