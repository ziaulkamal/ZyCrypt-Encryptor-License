package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "zycrypt",
	Short: "ZyCrypt — Ziya Encryptor License Manager",
	Long:  "ZyCrypt is a license management system for Laravel + Vue applications.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./zycrypt.yaml)")
}

func initConfig() {
	if err := config.Load(cfgFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}

func mustDB() {
	if err := db.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "Database connection failed: %v\n", err)
		os.Exit(1)
	}
}
