package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

func main() {
	fmt.Println(IntToBytes(24900))
	fmt.Println(BytesToInt([]byte{'2', '4', '9'}))
}
