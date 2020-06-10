package main

import "C"
import "fmt"

//go build -buildmode=c-shared -o /usr/local/lib/libcallback.so callback.go

//export Event
func Event() {
	fmt.Println("hello, C")
}

func main() {

}
