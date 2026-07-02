package main

import (
	"boilerplate-golang/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
