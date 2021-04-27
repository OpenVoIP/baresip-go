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
	"unsafe"

	log "github.com/sirupsen/logrus"
)

//UAConnect 拨打
func UAConnect(number string) {
	log.Printf("call %s", number)
	cNumber := C.CString(number)
	C.ua_connect(C.uag_current(), nil, nil, cNumber, C.VIDMODE_ON)
	defer C.free(unsafe.Pointer(cNumber))
}

//UAHangup 挂断
func UAHangup() {
	C.ua_hangup(C.uag_current(), nil, 0, nil)
}

//UAAnswer 接听
func UAAnswer() {
	/* Stop any ongoing ring-tones */
	// C.mem_deref(menu.play);
	C.ua_hold_answer(C.uag_current(), nil, C.VIDMODE_ON)
}
