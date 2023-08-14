package cmd

import (
	"fmt"
	"github.com/aurora/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type rootCmd struct {
	*baseCmd
}

func newRootCmd() *rootCmd {
	rc := &rootCmd{newBaseCmd()}
	rc.cmd = &cobra.Command{
		Use:   "aurora",
		Short: "A Golang command-line tool",
		// Long:    `A Golang command-line tool`,
		Version: consts.Version,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Usage(); err != nil {
				panic(err)
			}
		},
	}
	rc.cmd.SetUsageFunc(func(c *cobra.Command) error {
		if len(c.Commands()) > 0 {
			fmt.Printf("+--------------------------------------------------+\n|          ___                                     |\n|         /   | __  ___________  _________ _       |\n|        / /| |/ / / / ___/ __ \\/ ___/ __ `/       |\n|       / ___ / /_/ / /  / /_/ / /  / /_/ /        |\n|      /_/  |_\\__,_/_/   \\____/_/   \\__,_/         |\n|                                                  |\n|                                                  |\n|                >_ Aurora %s                  |\n|                   MIT License                    |\n|          Copyright (c) 2023 stubborn-gaga        |\n+--------------------------------------------------+\n\n", consts.Version)
		}
		fmt.Printf("[Command] %s\n", color.BlueString(c.Use))
		fmt.Printf("Flag Usages:\n%s \n", color.YellowString(c.Flags().FlagUsages()))
		if len(c.Commands()) > 0 {
			fmt.Printf("Command Usages:\n")
			for _, sc := range c.Commands() {
				if sc.Name() == "completion" || sc.Name() == "help" {
					continue
				}
				fmt.Printf("  %s %s\n", color.GreenString(sc.Name()), sc.Short)
			}
		}
		fmt.Println()
		return nil
	})
	rc.addCommands(
		newCreateCmd(),
		newInitCmd(),
		newGenModelCmd(),
		newBuildCmd(),
		newRunCmd(),
		newJobCmd(),
		//newCronCmd(),
	)

	return rc
}
