package cmd

import (
	"github.com/c16a/pouch/cli/dialer"
	"github.com/spf13/cobra"
	"log"
	"os"
	"os/signal"
)

var Seed string
var Url string
var ClientId string

func init() {
	connectCommand.Flags().StringVarP(&ClientId, "client-id", "c", "", "client id")
	connectCommand.Flags().StringVarP(&Seed, "seed", "s", "", "Seed for client authentication")
	connectCommand.Flags().StringVarP(&Url, "url", "u", "", "URL of Pouch server")
	rootCmd.AddCommand(connectCommand)
}

var connectCommand = &cobra.Command{
	Use:   "connect",
	Short: "Connect to a Pouch server",
	Run: func(cmd *cobra.Command, args []string) {

		encodedSeed := Seed

		dialer.DialTcp(Url, ClientId, encodedSeed)

		terminate := make(chan os.Signal, 1)
		signal.Notify(terminate, os.Interrupt, os.Kill)
		<-terminate
		log.Println("pouch exiting")
	},
}
