package main

import (
	"timebox-backend/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
