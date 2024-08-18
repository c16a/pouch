package cmd

import (
	"fmt"
	"github.com/c16a/pouch/sdk/pouchkey"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(genseedCmd)
}

var genseedCmd = &cobra.Command{
	Use:   "genseed",
	Short: "Creates a new seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		hexEncodedSeed, err := pouchkey.NewSeed()
		if err != nil {
			return err
		}
		fmt.Println(hexEncodedSeed)
		return nil
	},
}
