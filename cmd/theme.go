package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/theme"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage theme bundles",
}

var themeUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a theme ZIP bundle to the server",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		slug, _    := cmd.Flags().GetString("slug")
		name, _    := cmd.Flags().GetString("name")
		version, _ := cmd.Flags().GetString("version")
		file, _    := cmd.Flags().GetString("file")

		if slug == "" || name == "" || file == "" {
			fmt.Fprintln(os.Stderr, "Error: --slug, --name, and --file are required")
			os.Exit(1)
		}

		bundle, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		sum := sha256.Sum256(bundle)
		checksum := hex.EncodeToString(sum[:])

		t, err := svc.Upload(slug, name, version, bundle, checksum)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✓ Theme uploaded successfully")
		fmt.Printf("  ID       : %s\n", t.ID)
		fmt.Printf("  Slug     : %s\n", t.Slug)
		fmt.Printf("  Name     : %s\n", t.Name)
		fmt.Printf("  Version  : %s\n", t.Version)
		fmt.Printf("  Size     : %d bytes\n", t.FileSize)
		fmt.Printf("  Checksum : %s\n", t.Checksum)
	},
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all uploaded theme bundles",
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		themes, err := svc.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(themes) == 0 {
			fmt.Println("No themes uploaded yet.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "SLUG\tNAME\tVERSION\tSIZE")
		for _, t := range themes {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d bytes\n", t.Slug, t.Name, t.Version, t.FileSize)
		}
		w.Flush()
	},
}

var themeShowCmd = &cobra.Command{
	Use:   "show [slug]",
	Short: "Show details of a theme",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		t, err := svc.GetBySlug(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("  Slug      : %s\n", t.Slug)
		fmt.Printf("  Name      : %s\n", t.Name)
		fmt.Printf("  Version   : %s\n", t.Version)
		fmt.Printf("  Size      : %d bytes\n", t.FileSize)
		fmt.Printf("  Checksum  : %s\n", t.Checksum)
		fmt.Printf("  Uploaded  : %s\n", t.CreatedAt.Format("2006-01-02 15:04:05"))
	},
}

var themeDeleteCmd = &cobra.Command{
	Use:   "delete [slug]",
	Short: "Delete a theme bundle",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete theme %q? This cannot be undone. (yes/no): ", args[0])
			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Cancelled.")
				return
			}
		}

		if err := svc.Delete(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Theme %q deleted\n", args[0])
	},
}

var themeAssignCmd = &cobra.Command{
	Use:   "assign [theme-slug]",
	Short: "Assign a theme to a license plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		planSlug, _ := cmd.Flags().GetString("plan")
		if planSlug == "" {
			fmt.Fprintln(os.Stderr, "Error: --plan is required")
			os.Exit(1)
		}

		if err := svc.AssignToPlan(args[0], planSlug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Theme %q assigned to plan %q\n", args[0], planSlug)
	},
}

var themeUnassignCmd = &cobra.Command{
	Use:   "unassign [theme-slug]",
	Short: "Remove a theme from a license plan",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mustDB()
		svc := theme.NewService(db.DB, config.C.Security.SharedSecret)

		planSlug, _ := cmd.Flags().GetString("plan")
		if planSlug == "" {
			fmt.Fprintln(os.Stderr, "Error: --plan is required")
			os.Exit(1)
		}

		if err := svc.UnassignFromPlan(args[0], planSlug); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Theme %q removed from plan %q\n", args[0], planSlug)
	},
}

func init() {
	themeUploadCmd.Flags().String("slug", "", "Unique theme slug (required)")
	themeUploadCmd.Flags().String("name", "", "Display name (required)")
	themeUploadCmd.Flags().String("version", "1.0.0", "Version string")
	themeUploadCmd.Flags().String("file", "", "Path to ZIP file (required)")

	themeDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	themeAssignCmd.Flags().String("plan", "", "Plan slug to assign this theme to (required)")
	themeUnassignCmd.Flags().String("plan", "", "Plan slug to remove this theme from (required)")

	themeCmd.AddCommand(
		themeUploadCmd,
		themeListCmd,
		themeShowCmd,
		themeDeleteCmd,
		themeAssignCmd,
		themeUnassignCmd,
	)
	rootCmd.AddCommand(themeCmd)
}
