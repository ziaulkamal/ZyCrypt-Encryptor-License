package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/api"
	"github.com/ziaulkamal/zycrypt/config"
)

var servePort int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the ZyCrypt HTTP license server",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()

		if servePort > 0 {
			config.C.Server.Port = servePort
		}

		fmt.Printf("ZyCrypt server starting on %s\n", config.Addr())
		if err := api.Start(); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	},
}

func init() {
	serveCmd.Flags().IntVar(&servePort, "port", 0, "Override server port")
	rootCmd.AddCommand(serveCmd)
}
