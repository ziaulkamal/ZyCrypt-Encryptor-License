package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/db"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management",
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()

		fmt.Println("Running migrations...")
		if err := db.Migrate(); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Migrations complete")
	},
}

func init() {
	dbCmd.AddCommand(dbMigrateCmd)
	rootCmd.AddCommand(dbCmd)
}
