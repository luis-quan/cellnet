package binaryserial

import (
	"reflect"
)

func BinarySize(obj interface{}, alignMax int8) int {
	v := reflect.Indirect(reflect.ValueOf(obj))
	size := dataSize(v, v, alignMax)
	return size
}
