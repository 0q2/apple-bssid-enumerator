package cperm

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lcperm -Wl,-rpath,$ORIGIN
#include "cperm.h"
*/
import "C"
import "github.com/gigaryte/apple-bssid-enumerator/constants"

type CPerm struct {
	CPerm *C.struct_cperm_t
}

func (cp *CPerm) CreateCPerm(size uint, key [constants.KEYLEN]byte) {
	cp.CPerm = C.cperm_create(C.uint(size), C.PERM_MODE_PREFIX, C.PERM_CIPHER_RC5,
		(*C.uchar)(&key[0]), C.int(len(key)))
}
