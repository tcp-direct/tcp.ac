package main

import (
	"fmt"
	"github.com/twharmon/gouid"
)

func main() {
	s := gouid.String(8)
	fmt.Println(s)
}
