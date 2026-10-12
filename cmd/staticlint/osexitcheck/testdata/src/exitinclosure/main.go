package main

import "os"

func main() {
	defer func() {
		os.Exit(1)
	}()
}
