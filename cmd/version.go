package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// VersionCmd implements the 'version' command required by architect.
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Does nothing.",
		Run:   runVersionCmd,
	}
)

func runVersionCmd(cmd *cobra.Command, args []string) {
	fmt.Println("Not implemented.")
}
