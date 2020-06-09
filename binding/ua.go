package binding

/*
#cgo CFLAGS: -I/usr/local/include/re -I/usr/local/include/baresip
#cgo LDFLAGS: -ldl -lbaresip -lrem -lre

#include <stdint.h>
#include <stdlib.h>
#include <re.h>
#include <baresip.h>
*/
import "C"

import (
	"fmt"
	"unsafe"
)

//UAConnect 拨打
func UAConnect(number string) {
	fmt.Printf("call %s", number)
	cNumber := C.CString(number)
	C.ua_connect(C.uag_current(), nil, nil, cNumber, C.VIDMODE_ON)
	defer C.free(unsafe.Pointer(cNumber))
}
