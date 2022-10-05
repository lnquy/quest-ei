package hack

import (
	"encoding/binary"
	"reflect"
	"unsafe"
)

// StringToBytes converts a string to a byte slice.
//
// Note: This link the returned b slice to the s string,
// so any changes on b will also be reflected to s.
// Careful while using this function.
func StringToBytes(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

// BytesToString converts a string to a byte slice.
//
// Note: This link the returned b slice as the underlying value of the s string,
// so any changes on b will also be reflected to s.
// Make sure you won't change b or only changed it AFTER the returning s string has been used
// and it won't be accessed anymore.
// Careful while using this function.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(i))
	return b
}

func BytesToInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
