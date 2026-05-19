package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	zcrypto "github.com/ziaulkamal/zycrypt/crypto"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

var keyCmd = &cobra.Command{
	Use:   "key",
	Short: "Key utilities",
}

var keyGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Preview a key without saving to DB",
	Run: func(cmd *cobra.Command, args []string) {
		planSlug, _ := cmd.Flags().GetString("plan")
		years, _ := cmd.Flags().GetInt("years")

		typeChar := "S"
		if planSlug != "" {
			switch {
			case years == 0:
				typeChar = "L"
			default:
				typeChar = "S"
			}
		}

		// If DB available, resolve plan type from slug
		if planSlug != "" {
			if err := config.Load(cfgFile); err == nil {
				mustDB()
				svc := license.NewService(nil, config.C.Security.SharedSecret)
				_ = svc
			}
		}

		key, err := zcrypto.GenerateKey(typeChar, years, config.C.Security.SharedSecret)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Preview key (not saved to DB):")
		fmt.Printf("  Key  : %s\n", key)
		fmt.Printf("  Type : %s\n", typeChar)
		fmt.Printf("  Years: %d (0=lifetime)\n", years)
	},
}

var keyVerifyCmd = &cobra.Command{
	Use:   "verify [key]",
	Short: "Verify key format and checksum",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		if err := zcrypto.VerifyKey(key, config.C.Security.SharedSecret); err != nil {
			fmt.Printf("✗ Invalid: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Key is valid\n")
		fmt.Printf("  Type     : %s\n", zcrypto.PlanTypeFromKey(key))
		fmt.Printf("  Duration : %d years (0=lifetime)\n", zcrypto.DurationFromKey(key))
	},
}

func init() {
	keyGenerateCmd.Flags().String("plan", "", "Plan slug (for type encoding)")
	keyGenerateCmd.Flags().Int("years", 1, "Duration years (0=lifetime)")

	keyCmd.AddCommand(keyGenerateCmd, keyVerifyCmd)
	rootCmd.AddCommand(keyCmd)
}
