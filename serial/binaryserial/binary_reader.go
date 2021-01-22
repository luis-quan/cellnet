package binaryserial

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

func BinaryRead(data []byte, obj interface{}, alignMax int8) error {

	if len(data) == 0 {
		return nil
	}

	v := reflect.ValueOf(obj)

	switch v.Kind() {
	case reflect.Ptr:
		v = v.Elem()
	}

	size := dataSize(v, v, alignMax)
	if size < 0 {
		return ErrInvalidType
	}

	if len(data) < size {
		return ErrOutOfData
	}

	fmt.Println(len(data), size)

	d := &decoder{order: binary.LittleEndian, buf: data}
	d.value(v, v, alignMax)

	return nil
}
