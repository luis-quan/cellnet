package binaryserial

import (
	"math"
	"reflect"
	"sync"
)

// A ByteOrder specifies how to convert byte sequences into
// 16-, 32-, or 64-bit unsigned integers.
type ByteOrder interface {
	Uint16([]byte) uint16
	Uint32([]byte) uint32
	Uint64([]byte) uint64
	PutUint16([]byte, uint16)
	PutUint32([]byte, uint32)
	PutUint64([]byte, uint64)
	String() string
}

// LittleEndian is the little-endian implementation of ByteOrder.
var LittleEndian littleEndian

// BigEndian is the big-endian implementation of ByteOrder.
var BigEndian bigEndian

type littleEndian struct{}

func (littleEndian) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

func (littleEndian) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func (littleEndian) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (littleEndian) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (littleEndian) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func (littleEndian) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

func (littleEndian) String() string { return "LittleEndian" }

func (littleEndian) GoString() string { return "binary.LittleEndian" }

type bigEndian struct{}

func (bigEndian) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[1]) | uint16(b[0])<<8
}

func (bigEndian) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func (bigEndian) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[3]) | uint32(b[2])<<8 | uint32(b[1])<<16 | uint32(b[0])<<24
}

func (bigEndian) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 24)
	b[1] = byte(v >> 16)
	b[2] = byte(v >> 8)
	b[3] = byte(v)
}

func (bigEndian) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[7]) | uint64(b[6])<<8 | uint64(b[5])<<16 | uint64(b[4])<<24 |
		uint64(b[3])<<32 | uint64(b[2])<<40 | uint64(b[1])<<48 | uint64(b[0])<<56
}

func (bigEndian) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 56)
	b[1] = byte(v >> 48)
	b[2] = byte(v >> 40)
	b[3] = byte(v >> 32)
	b[4] = byte(v >> 24)
	b[5] = byte(v >> 16)
	b[6] = byte(v >> 8)
	b[7] = byte(v)
}

func (bigEndian) String() string { return "BigEndian" }

func (bigEndian) GoString() string { return "binary.BigEndian" }

// Size returns how many bytes Write would generate to encode the value v, which
// must be a fixed-size value or a slice of fixed-size values, or a pointer to such data.
// If v is neither of these, Size returns -1.
func Size(v interface{}, alignMax int8) int {
	vv := reflect.Indirect(reflect.ValueOf(v))
	return dataSize(vv, vv, alignMax)
}

var structSize sync.Map // map[reflect.Type]int

//对齐方法
func alignFunc(dataSize int, align int8) int {
	size := int8(dataSize % int(align))
	if size > 0 {
		dataSize += int(align - size)
	}
	return dataSize
}

//获取该类型的对齐大小
func alignSize(t reflect.Type, alignMax int8) int8 {
	var align int8 = 1

	switch t.Kind() {
	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		align = int8(t.Size())
	case reflect.Struct:
		var a int8 = 1
		for i, n := 0, t.NumField(); i < n; i++ {
			a = alignSize(t.Field(i).Type, alignMax)
			if align < a {
				align = a
			}
		}
	case reflect.Slice, reflect.Array:
		align = alignSize(t.Elem(), alignMax)
	}

	if align > alignMax {
		align = alignMax
	}

	return align
}

// dataSize returns the number of bytes the actual data represented by v occupies in memory.
// For compound structures, it sums the sizes of the elements. Thus, for instance, for a slice
// it returns the length of the slice times the element size and does not count the memory
// occupied by the header. If the type of v is not acceptable, dataSize returns -1.
// alignMax对齐补齐

/*
	type S struct {
		SliceLen int `slicefrom:SliceName`
		SliceName []int
	}
*/

func getSliceLen(v reflect.Value, slicename string) int {
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i, n := 0, t.NumField(); i < n; i++ {
			field := t.Field(i)
			slicefrom := field.Tag.Get("slicefrom")
			if slicefrom == slicename {
				switch field.Type.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					return int(v.Field(i).Int())
				case reflect.Struct:
					return getSliceLen(v.Field(i), slicename)
				}
			}
		}
	}

	return 0
}

func dataSize(ov reflect.Value, v reflect.Value, alignMax int8) int {
	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		size := 0
		if l > 0 {
			size = dataSize(ov, v.Index(0), alignMax)
		}

		align := alignSize(v.Type(), alignMax)
		size = alignFunc(size, align)
		return size * l
	case reflect.Slice:
		l := getSliceLen(ov, v.Type().Name())
		size := 0
		if l > 0 {
			if v.Len() <= 0 {
				//创建一个大小的
				vv := reflect.New(v.Type().Elem())
				vv = reflect.Indirect(vv)
				size = dataSize(v, vv, alignMax)
			} else {
				size = dataSize(v, v.Index(0), alignMax)
			}
		}

		align := alignSize(v.Type(), alignMax)
		size = alignFunc(size, align)
		return size * l
	case reflect.Struct:
		t := v.Type()
		// if size, ok := structSize.Load(t); ok {
		// 	return size.(int)
		// }
		size := 0
		for i, n := 0, t.NumField(); i < n; i++ {
			s := dataSize(v, v.Field(i), alignMax)
			if s < 0 {
				return -1
			}
			a := alignSize(v.Field(i).Type(), alignMax)
			size = alignFunc(size, a)
			size += s
		}
		align := alignSize(t, alignMax)
		size = alignFunc(size, align)
		//structSize.Store(t, size)
		//结构体需要补齐
		return size
	default:
		size := sizeof(v.Type())
		return size
	}
}

// sizeof returns the size >= 0 of variables for the given type or -1 if the type is not acceptable.
func sizeof(t reflect.Type) int {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return int(t.Size())
	}

	return -1
}

type coder struct {
	order  ByteOrder
	buf    []byte
	offset int
}

type decoder coder
type encoder coder

func (d *decoder) bool() bool {
	x := d.buf[d.offset]
	d.offset++
	return x != 0
}

func (e *encoder) bool(x bool) {
	if x {
		e.buf[e.offset] = 1
	} else {
		e.buf[e.offset] = 0
	}
	e.offset++
}

func (d *decoder) uint8() uint8 {
	x := d.buf[d.offset]
	d.offset++
	return x
}

func (e *encoder) uint8(x uint8) {
	e.buf[e.offset] = x
	e.offset++
}

func (d *decoder) uint16() uint16 {
	x := d.order.Uint16(d.buf[d.offset : d.offset+2])
	d.offset += 2
	return x
}

func (e *encoder) uint16(x uint16) {
	e.order.PutUint16(e.buf[e.offset:e.offset+2], x)
	e.offset += 2
}

func (d *decoder) uint32() uint32 {
	x := d.order.Uint32(d.buf[d.offset : d.offset+4])
	d.offset += 4
	return x
}

func (e *encoder) uint32(x uint32) {
	e.order.PutUint32(e.buf[e.offset:e.offset+4], x)
	e.offset += 4
}

func (d *decoder) uint64() uint64 {
	x := d.order.Uint64(d.buf[d.offset : d.offset+8])
	d.offset += 8
	return x
}

func (e *encoder) uint64(x uint64) {
	e.order.PutUint64(e.buf[e.offset:e.offset+8], x)
	e.offset += 8
}

func (d *decoder) int8() int8 { return int8(d.uint8()) }

func (e *encoder) int8(x int8) { e.uint8(uint8(x)) }

func (d *decoder) int16() int16 { return int16(d.uint16()) }

func (e *encoder) int16(x int16) { e.uint16(uint16(x)) }

func (d *decoder) int32() int32 { return int32(d.uint32()) }

func (e *encoder) int32(x int32) { e.uint32(uint32(x)) }

func (d *decoder) int64() int64 { return int64(d.uint64()) }

func (e *encoder) int64(x int64) { e.uint64(uint64(x)) }

func (d *decoder) valueslice(l int, v reflect.Value, alignMax int8) {
	for i := 0; i < l; i++ {
		if i >= v.Len() {
			//插入新的对象
			v1 := reflect.Append(v, reflect.New(v.Type().Elem()).Elem())
			v.Set(v1)
			d.value(v, v.Index(i), alignMax)
		} else {
			d.value(v, v.Index(i), alignMax)
		}
	}
}

func (d *decoder) value(ov reflect.Value, v reflect.Value, alignMax int8) {
	align := alignSize(v.Type(), alignMax)
	d.offset = alignFunc(d.offset, align)

	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			d.value(v, v.Index(i), alignMax)
		}
	case reflect.Struct:
		t := v.Type()
		l := v.NumField()
		/*
			type ss struct {
				aa []int
			}
		*/
		//如果第一次 并且只有一个切片 就先获取单个切片的大小切片 更加大小来取
		if ov == v && l == 1 && t.Field(0).Type.Kind() == reflect.Slice {
			v = v.Field(0)
			if v.CanSet() || t.Field(0).Name != "_" {
				vv := reflect.New(v.Type().Elem())
				vv = reflect.Indirect(vv)
				size := dataSize(v, vv, alignMax)
				l = len(d.buf) / size
				d.valueslice(l, v, alignMax)
			}
		} else {
			ov = v
			for i := 0; i < l; i++ {
				// Note: Calling v.CanSet() below is an optimization.
				// It would be sufficient to check the field name,
				// but creating the StructField info for each field is
				// costly (run "go test -bench=ReadStruct" and compare
				// results when making changes to this code).
				if v := v.Field(i); v.CanSet() || t.Field(i).Name != "_" {
					d.value(ov, v, alignMax)
				} else {
					d.skip(ov, v, alignMax)
				}
			}
		}
	case reflect.Slice:
		l := getSliceLen(ov, v.Type().Name())
		d.valueslice(l, v, alignMax)
	case reflect.Bool:
		v.SetBool(d.bool())
	case reflect.Int8:
		v.SetInt(int64(d.int8()))
	case reflect.Int16:
		v.SetInt(int64(d.int16()))
	case reflect.Int32:
		v.SetInt(int64(d.int32()))
	case reflect.Int64:
		v.SetInt(d.int64())

	case reflect.Uint8:
		v.SetUint(uint64(d.uint8()))
	case reflect.Uint16:
		v.SetUint(uint64(d.uint16()))
	case reflect.Uint32:
		v.SetUint(uint64(d.uint32()))
	case reflect.Uint64:
		v.SetUint(d.uint64())

	case reflect.Float32:
		v.SetFloat(float64(math.Float32frombits(d.uint32())))
	case reflect.Float64:
		v.SetFloat(math.Float64frombits(d.uint64()))

	case reflect.Complex64:
		v.SetComplex(complex(
			float64(math.Float32frombits(d.uint32())),
			float64(math.Float32frombits(d.uint32())),
		))
	case reflect.Complex128:
		v.SetComplex(complex(
			math.Float64frombits(d.uint64()),
			math.Float64frombits(d.uint64()),
		))
	}
}

func (e *encoder) valueslice(l int, v reflect.Value, alignMax int8) {
	for i := 0; i < l; i++ {
		if i >= v.Len() {
			//插入新的对象
			v1 := reflect.Append(v, reflect.New(v.Type().Elem()).Elem())
			v.Set(v1)
			e.value(v, v.Index(i), alignMax)
		} else {
			e.value(v, v.Index(i), alignMax)
		}
	}
}

func (e *encoder) value(ov reflect.Value, v reflect.Value, alignMax int8) {
	align := alignSize(v.Type(), alignMax)
	e.offset = alignFunc(e.offset, align)

	switch v.Kind() {
	case reflect.Array:
		l := v.Len()
		for i := 0; i < l; i++ {
			e.value(v, v.Index(i), alignMax)
		}

	case reflect.Struct:
		t := v.Type()
		l := v.NumField()

		//如果第一次 并且只有一个切片 就先获取单个切片的大小切片 更加大小来取
		if ov == v && l == 1 && t.Field(0).Type.Kind() == reflect.Slice {
			v = v.Field(0)
			if v.CanSet() || t.Field(0).Name != "_" {
				vv := reflect.New(v.Type().Elem())
				vv = reflect.Indirect(vv)
				size := dataSize(v, vv, alignMax)
				l = len(e.buf) / size
				e.valueslice(l, v, alignMax)
			}
		} else {
			for i := 0; i < l; i++ {
				// see comment for corresponding code in decoder.value()
				if v := v.Field(i); v.CanSet() || t.Field(i).Name != "_" {
					e.value(ov, v, alignMax)
				} else {
					e.skip(ov, v, alignMax)
				}
			}
		}

	case reflect.Slice:
		l := getSliceLen(ov, v.Type().Name())
		e.valueslice(l, v, alignMax)

	case reflect.Bool:
		e.bool(v.Bool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch v.Type().Kind() {
		case reflect.Int8:
			e.int8(int8(v.Int()))
		case reflect.Int16:
			e.int16(int16(v.Int()))
		case reflect.Int32:
			e.int32(int32(v.Int()))
		case reflect.Int64:
			e.int64(v.Int())
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		switch v.Type().Kind() {
		case reflect.Uint8:
			e.uint8(uint8(v.Uint()))
		case reflect.Uint16:
			e.uint16(uint16(v.Uint()))
		case reflect.Uint32:
			e.uint32(uint32(v.Uint()))
		case reflect.Uint64:
			e.uint64(v.Uint())
		}

	case reflect.Float32, reflect.Float64:
		switch v.Type().Kind() {
		case reflect.Float32:
			e.uint32(math.Float32bits(float32(v.Float())))
		case reflect.Float64:
			e.uint64(math.Float64bits(v.Float()))
		}

	case reflect.Complex64, reflect.Complex128:
		switch v.Type().Kind() {
		case reflect.Complex64:
			x := v.Complex()
			e.uint32(math.Float32bits(float32(real(x))))
			e.uint32(math.Float32bits(float32(imag(x))))
		case reflect.Complex128:
			x := v.Complex()
			e.uint64(math.Float64bits(real(x)))
			e.uint64(math.Float64bits(imag(x)))
		}
	}
}

func (d *decoder) skip(ov reflect.Value, v reflect.Value, alignMax int8) {
	size := dataSize(ov, v, alignMax)
	d.offset += size
}

func (e *encoder) skip(ov reflect.Value, v reflect.Value, alignMax int8) {
	n := dataSize(ov, v, alignMax)
	zero := e.buf[e.offset : e.offset+n]
	for i := range zero {
		zero[i] = 0
	}
	e.offset += n
}

// intDataSize returns the size of the data required to represent the data when encoded.
// It returns zero if the type cannot be implemented by the fast path in Read or Write.
func intDataSize(data interface{}) int {
	switch data := data.(type) {
	case bool, int8, uint8, *bool, *int8, *uint8:
		return 1
	case []bool:
		return len(data)
	case []int8:
		return len(data)
	case []uint8:
		return len(data)
	case int16, uint16, *int16, *uint16:
		return 2
	case []int16:
		return 2 * len(data)
	case []uint16:
		return 2 * len(data)
	case int32, uint32, *int32, *uint32:
		return 4
	case []int32:
		return 4 * len(data)
	case []uint32:
		return 4 * len(data)
	case int64, uint64, *int64, *uint64:
		return 8
	case []int64:
		return 8 * len(data)
	case []uint64:
		return 8 * len(data)
	case float32, *float32:
		return 4
	case float64, *float64:
		return 8
	case []float32:
		return 4 * len(data)
	case []float64:
		return 8 * len(data)
	}
	return 0
}
