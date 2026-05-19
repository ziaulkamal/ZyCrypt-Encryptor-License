package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Manage licenses",
}

var licenseCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new license",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		name, _ := cmd.Flags().GetString("name")
		email, _ := cmd.Flags().GetString("email")
		plan, _ := cmd.Flags().GetString("plan")

		if name == "" || email == "" || plan == "" {
			fmt.Fprintln(os.Stderr, "Error: --name, --email, and --plan are required")
			os.Exit(1)
		}

		lic, err := svc.Create(name, email, plan)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ License created successfully")
		fmt.Printf("  Key     : %s\n", lic.LicenseKey)
		fmt.Printf("  Plan    : %s\n", plan)
		if lic.IsLifetime {
			fmt.Println("  Expired : never (lifetime)")
		} else if lic.ExpiresAt != nil {
			fmt.Printf("  Expired : %s\n", lic.ExpiresAt.Format("2006-01-02"))
		}
		fmt.Printf("  Status  : %s\n", lic.Status)
	},
}

var licenseListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all licenses",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		status, _ := cmd.Flags().GetString("status")
		licenses, err := svc.List(status)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "KEY\tCUSTOMER\tPLAN\tSTATUS\tEXPIRES")
		for _, l := range licenses {
			exp := "lifetime"
			if !l.IsLifetime && l.ExpiresAt != nil {
				exp = l.ExpiresAt.Format("2006-01-02")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				l.LicenseKey, l.CustomerName, l.Plan.Slug, l.Status, exp)
		}
		w.Flush()
	},
}

var licenseShowCmd = &cobra.Command{
	Use:   "show [key]",
	Short: "Show full license details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		lic, err := svc.GetByKey(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "License not found: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Key           : %s\n", lic.LicenseKey)
		fmt.Printf("Customer      : %s <%s>\n", lic.CustomerName, lic.CustomerEmail)
		fmt.Printf("Plan          : %s (limit: %d)\n", lic.Plan.Slug, lic.Plan.SiteLimit)
		fmt.Printf("Status        : %s\n", lic.Status)
		if lic.BannedReason != nil {
			fmt.Printf("Ban reason    : %s\n", *lic.BannedReason)
		}
		fmt.Printf("Purchase date : %s\n", lic.PurchaseDate.Format("2006-01-02"))
		if lic.IsLifetime {
			fmt.Println("Expires       : never (lifetime)")
		} else if lic.ExpiresAt != nil {
			fmt.Printf("Expires       : %s\n", lic.ExpiresAt.Format("2006-01-02"))
		}

		fmt.Println("\nDomains:")
		if len(lic.Domains) == 0 {
			fmt.Println("  (none)")
		}
		for _, d := range lic.Domains {
			primary := ""
			if d.IsPrimary {
				primary = " [primary]"
			}
			fmt.Printf("  - %s%s (since %s)\n", d.Domain, primary, d.RegisteredAt.Format("2006-01-02"))
		}

		logs, _ := svc.RecentLogs(lic.ID, 10)
		fmt.Println("\nRecent events:")
		if len(logs) == 0 {
			fmt.Println("  (none)")
		}
		for _, l := range logs {
			fmt.Printf("  [%s] %s\n", l.CreatedAt.Format("2006-01-02 15:04:05"), l.Event)
		}
	},
}

var licenseBanCmd = &cobra.Command{
	Use:   "ban [key]",
	Short: "Ban a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		reason, _ := cmd.Flags().GetString("reason")
		if reason == "" {
			fmt.Fprintln(os.Stderr, "Error: --reason is required")
			os.Exit(1)
		}

		if err := svc.Ban(args[0], reason); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ License %s banned\n", args[0])
	},
}

var licenseUnbanCmd = &cobra.Command{
	Use:   "unban [key]",
	Short: "Unban a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		if err := svc.Unban(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ License %s unbanned\n", args[0])
	},
}

var licenseExtendCmd = &cobra.Command{
	Use:   "extend [key]",
	Short: "Extend license expiry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		years, _ := cmd.Flags().GetInt("years")
		if years <= 0 {
			years = 1
		}

		if err := svc.Extend(args[0], years); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ License %s extended by %d year(s)\n", args[0], years)
	},
}

var licenseRevokeCmd = &cobra.Command{
	Use:   "revoke [key]",
	Short: "Permanently delete a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		fmt.Printf("WARNING: This will permanently delete license %s and all associated data.\n", args[0])
		fmt.Print("Type the license key to confirm: ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		confirm := strings.TrimSpace(scanner.Text())

		if confirm != args[0] {
			fmt.Println("Aborted.")
			os.Exit(0)
		}

		if err := svc.Revoke(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ License %s revoked and deleted\n", args[0])
	},
}

func init() {
	licenseCreateCmd.Flags().String("name", "", "Customer name (required)")
	licenseCreateCmd.Flags().String("email", "", "Customer email (required)")
	licenseCreateCmd.Flags().String("plan", "", "Plan slug (required)")

	licenseListCmd.Flags().String("status", "", "Filter by status: active|inactive|banned|expired")

	licenseBanCmd.Flags().String("reason", "", "Ban reason (required)")

	licenseExtendCmd.Flags().Int("years", 1, "Number of years to extend")

	licenseCmd.AddCommand(
		licenseCreateCmd,
		licenseListCmd,
		licenseShowCmd,
		licenseBanCmd,
		licenseUnbanCmd,
		licenseExtendCmd,
		licenseRevokeCmd,
	)
	rootCmd.AddCommand(licenseCmd)
}
