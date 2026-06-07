package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "folder-exposer",
	Short: "A Zero-Trust Network Access Gateway",
	Long:  `FolderExposer allows you to securely expose local directories to the public internet using custom raw TCP socket proxying and HTTP protocol hijacking.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
