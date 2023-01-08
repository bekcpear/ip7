package main

import (
	"ip7/internal/cmd/root"
	"os"
)

func main() {
	var ret int
	cmd := root.NewRootCmd()
	err := cmd.Execute()
	if err != nil {
		ret = 1
	}
	os.Exit(ret)
}
