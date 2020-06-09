package main

import (
	"time"

	"github.com/OpenVoIP/baresip-go/binding"
)

func main() {
	go testDial()

	binding.Start()

}

func testDial() {
	time.Sleep(5 * time.Second)
	binding.UAConnect("*61")
}
