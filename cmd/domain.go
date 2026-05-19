package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/domain"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

var domainCmd = &cobra.Command{
	Use:   "domain",
	Short: "Manage license domains",
}

var domainAddCmd = &cobra.Command{
	Use:   "add [key]",
	Short: "Add a domain to a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		licSvc := license.NewService(db.DB, config.C.Security.SharedSecret)
		domSvc := domain.NewService(db.DB)

		domainName, _ := cmd.Flags().GetString("domain")
		if domainName == "" {
			fmt.Fprintln(os.Stderr, "Error: --domain is required")
			os.Exit(1)
		}

		lic, err := licSvc.GetByKey(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "License not found: %v\n", err)
			os.Exit(1)
		}

		if err := domSvc.Add(lic.ID, domainName, lic.Plan.SiteLimit); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Domain %q added to license %s\n", domainName, args[0])
	},
}

var domainRemoveCmd = &cobra.Command{
	Use:   "remove [key]",
	Short: "Remove a domain from a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		licSvc := license.NewService(db.DB, config.C.Security.SharedSecret)
		domSvc := domain.NewService(db.DB)

		domainName, _ := cmd.Flags().GetString("domain")
		if domainName == "" {
			fmt.Fprintln(os.Stderr, "Error: --domain is required")
			os.Exit(1)
		}

		lic, err := licSvc.GetByKey(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "License not found: %v\n", err)
			os.Exit(1)
		}

		if err := domSvc.Remove(lic.ID, domainName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Domain %q removed from license %s\n", domainName, args[0])
	},
}

var domainListCmd = &cobra.Command{
	Use:   "list [key]",
	Short: "List domains for a license",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		licSvc := license.NewService(db.DB, config.C.Security.SharedSecret)
		domSvc := domain.NewService(db.DB)

		lic, err := licSvc.GetByKey(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "License not found: %v\n", err)
			os.Exit(1)
		}

		domains, err := domSvc.List(lic.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(domains) == 0 {
			fmt.Println("No domains registered.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "DOMAIN\tPRIMARY\tREGISTERED")
		for _, d := range domains {
			primary := "no"
			if d.IsPrimary {
				primary = "yes"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", d.Domain, primary, d.RegisteredAt.Format("2006-01-02"))
		}
		w.Flush()
	},
}

func init() {
	domainAddCmd.Flags().String("domain", "", "Hostname without https:// (required)")
	domainRemoveCmd.Flags().String("domain", "", "Hostname to remove (required)")

	domainCmd.AddCommand(domainAddCmd, domainRemoveCmd, domainListCmd)
	rootCmd.AddCommand(domainCmd)
}
