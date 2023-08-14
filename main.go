package main

import (
	"fmt"
	"github.com/aurora/cmd"
	"os"
)

func main() {
	app := cmd.NewCommand()
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
