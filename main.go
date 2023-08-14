package main

import (
	"fmt"
	"github.com/stubborn-gaga-0805/aurora/cmd"
	"os"
)

func main() {
	app := cmd.NewCommand()
	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
