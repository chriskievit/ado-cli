/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var WorkItemId int

// linkCmd represents the link command
var linkCmd = &cobra.Command{
	Use:   "link",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("link called for %d\n", WorkItemId)
	},
}

func init() {
	rootCmd.AddCommand(linkCmd)

	linkCmd.Flags().IntVarP(&WorkItemId, "work-item", "w", 0, "Work item ID to link to branch")
	linkCmd.MarkFlagRequired("work-item")
}
