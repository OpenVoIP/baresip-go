package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	var sourceStr string = `
	--- User Agents (2) ---
> sip:1004@192.168.11.242                    OK  
  sip:1005@192.168.11.242                    ERR 

	`
	fmt.Printf("value %t", strings.Contains(sourceStr, "User Agents"))

	re := regexp.MustCompile(`sip:(\d+)@\S+\w+\s+(\w+)`)
	matched := re.FindAllStringSubmatch(sourceStr, -1)
	for _, match := range matched {
		fmt.Printf("exten is: %s, status is: %s\n", match[1], match[2])
	}
}
