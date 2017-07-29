package main

import (
	"fmt"
	"os"
	"text/scanner"
)

func main() {
	var s scanner.Scanner
	s.Init(os.Stdin)
	for {
		fmt.Print("> ")
		x := s.Scan()
		fmt.Println(x, s.TokenText())
	}
}
