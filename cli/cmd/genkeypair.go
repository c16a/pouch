package cmd

import (
	"encoding/hex"
	"fmt"
	"github.com/c16a/pouch/sdk/pouchkey"
	"github.com/spf13/cobra"
)

func init() {
	genkeypairCmd.Flags().StringVarP(&Seed, "seed", "s", "", "seed for key pair")
	rootCmd.AddCommand(genkeypairCmd)
}

var genkeypairCmd = &cobra.Command{
	Use:   "genkeypair",
	Short: "Creates a new Ed448 keypair",
	RunE: func(cmd *cobra.Command, args []string) error {
		encodedSeed := Seed

		seed, err := hex.DecodeString(encodedSeed)
		if err != nil {
			return err
		}

		privateKey, publicKey := pouchkey.NewHexKeys(seed)
		fmt.Printf("Private Key: %s\n", privateKey)
		fmt.Printf("Public Key: %s\n", publicKey)

		return nil
	},
}
