package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Manage license plans",
}

var planCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new license plan",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		limit, _ := cmd.Flags().GetInt("limit")
		duration, _ := cmd.Flags().GetString("duration")
		price, _ := cmd.Flags().GetFloat64("price")

		if name == "" || slug == "" {
			fmt.Fprintln(os.Stderr, "Error: --name and --slug are required")
			os.Exit(1)
		}

		plan, err := svc.CreatePlan(name, slug, limit, duration, price)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Plan created successfully")
		fmt.Printf("  ID       : %s\n", plan.ID)
		fmt.Printf("  Name     : %s\n", plan.Name)
		fmt.Printf("  Slug     : %s\n", plan.Slug)
		fmt.Printf("  Limit    : %d\n", plan.SiteLimit)
		fmt.Printf("  Duration : %s\n", plan.Duration)
		fmt.Printf("  Price    : %.2f\n", plan.Price)
	},
}

var planListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all license plans",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		plans, err := svc.ListPlans()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SLUG\tNAME\tLIMIT\tDURATION\tPRICE\tACTIVE")
		for _, p := range plans {
			active := "yes"
			if !p.IsActive {
				active = "no"
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%.2f\t%s\n",
				p.Slug, p.Name, p.SiteLimit, p.Duration, p.Price, active)
		}
		w.Flush()
	},
}

var planUpdateCmd = &cobra.Command{
	Use:   "update [slug]",
	Short: "Update plan attributes",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		updates := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			updates["name"] = v
		}
		if v, _ := cmd.Flags().GetFloat64("price"); v > 0 {
			updates["price"] = v
		}
		if v, _ := cmd.Flags().GetInt("limit"); v != 0 {
			updates["site_limit"] = v
		}

		if len(updates) == 0 {
			fmt.Fprintln(os.Stderr, "No fields to update. Use --name, --price, or --limit")
			os.Exit(1)
		}

		if err := svc.UpdatePlan(args[0], updates); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Plan %q updated\n", args[0])
	},
}

var planDisableCmd = &cobra.Command{
	Use:   "disable [slug]",
	Short: "Disable a plan (existing licenses unaffected)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := license.NewService(db.DB, config.C.Security.SharedSecret)

		if err := svc.DisablePlan(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Plan %q disabled\n", args[0])
	},
}

func init() {
	planCreateCmd.Flags().String("name", "", "Display name (required)")
	planCreateCmd.Flags().String("slug", "", "Unique slug (required)")
	planCreateCmd.Flags().Int("limit", 1, "Domain limit: 1 | 5 | -1 (unlimited)")
	planCreateCmd.Flags().String("duration", "1y", "Duration: 1y | 3y | lifetime")
	planCreateCmd.Flags().Float64("price", 0, "Price in Rupiah")

	planUpdateCmd.Flags().String("name", "", "New display name")
	planUpdateCmd.Flags().Float64("price", 0, "New price")
	planUpdateCmd.Flags().Int("limit", 0, "New site limit")

	planCmd.AddCommand(planCreateCmd, planListCmd, planUpdateCmd, planDisableCmd)
	rootCmd.AddCommand(planCmd)
}
