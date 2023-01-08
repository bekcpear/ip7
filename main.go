package main

import (
	"os"

	"oss.ac/ip7/internal/cmd/root"
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
