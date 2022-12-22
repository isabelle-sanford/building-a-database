package main

import (
	"fmt"
	"os"
	"strconv"
)

var BLOCKSIZE int = os.Getpagesize()


func main() {
	fmt.Println(BLOCKSIZE)
	fmt.Println(strconv.IntSize)
}