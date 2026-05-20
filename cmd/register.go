package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ziaulkamal/zycrypt/config"
	"github.com/ziaulkamal/zycrypt/db"
	"github.com/ziaulkamal/zycrypt/internal/license"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Buat lisensi baru dengan panduan langkah demi langkah",
	Run:   runRegister,
}

func init() {
	rootCmd.AddCommand(registerCmd)
}

func runRegister(cmd *cobra.Command, args []string) {
	mustDB()

	scanner := bufio.NewScanner(os.Stdin)
	licSvc := license.NewService(db.DB, config.C.Security.SharedSecret)

	printHeader()

	// ── Step 1: Tipe paket ──────────────────────────────────────────────
	printStep(1, "Pilih Paket")
	fmt.Println("   1. Single Site  — 1 domain")
	fmt.Println("   2. Multi Site   — hingga 5 domain")
	paketPilihan := promptChoice(scanner, "   Pilihan [1/2]", []string{"1", "2"})

	siteLimit := 1
	paketLabel := "Single Site"
	if paketPilihan == "2" {
		siteLimit = 5
		paketLabel = "Multi Site"
	}

	// ── Step 2: Durasi ──────────────────────────────────────────────────
	printStep(2, "Pilih Durasi")
	fmt.Println("   1. 1 Tahun")
	fmt.Println("   2. 3 Tahun")
	fmt.Println("   3. Lifetime (tanpa batas)")
	durasiPilihan := promptChoice(scanner, "   Pilihan [1/2/3]", []string{"1", "2", "3"})

	durationStr := "1y"
	durasiLabel := "1 Tahun"
	switch durasiPilihan {
	case "2":
		durationStr = "3y"
		durasiLabel = "3 Tahun"
	case "3":
		durationStr = "lifetime"
		durasiLabel = "Lifetime"
	}

	// ── Step 3: Identifikasi pelanggan ──────────────────────────────────
	printStep(3, "Identifikasi Pelanggan")
	customerName := prompt(scanner, "   Nama pelanggan")
	if customerName == "" {
		customerName = "Customer"
	}
	customerEmail := prompt(scanner, "   Email pelanggan")
	if customerEmail == "" {
		customerEmail = "noreply@example.com"
	}

	// ── Proses ──────────────────────────────────────────────────────────
	fmt.Println()
	fmt.Println("  Memproses...")

	plan, err := findOrCreatePlan(licSvc, siteLimit, durationStr, paketLabel)
	if err != nil {
		fatal("Gagal menyiapkan plan: " + err.Error())
	}

	lic, err := licSvc.Create(customerName, customerEmail, plan.Slug)
	if err != nil {
		fatal("Gagal membuat lisensi: " + err.Error())
	}

	// ── Hasil ───────────────────────────────────────────────────────────
	fmt.Println()
	printDivider()
	fmt.Println("  ✓  Lisensi Berhasil Dibuat!")
	printDivider()
	fmt.Printf("  Pelanggan : %s (%s)\n", customerName, customerEmail)
	fmt.Printf("  Key       : %s\n", lic.LicenseKey)
	fmt.Printf("  Paket     : %s\n", paketLabel)
	fmt.Printf("  Durasi    : %s\n", durasiLabel)
	if lic.IsLifetime {
		fmt.Println("  Expired   : Tidak pernah (Lifetime)")
	} else if lic.ExpiresAt != nil {
		fmt.Printf("  Expired   : %s\n", lic.ExpiresAt.Format("02 January 2006"))
	}
	fmt.Printf("  Dibuat    : %s\n", time.Now().Format("02 January 2006, 15:04"))
	printDivider()
	fmt.Println()
	fmt.Println("  Berikan key ini ke client, lalu minta mereka set di .env:")
	fmt.Printf("  ZYCRYPT_LICENSE_KEY=%s\n", lic.LicenseKey)
	fmt.Printf("  ZYCRYPT_SERVER_URL=%s\n", config.C.Server.BaseURL)
	fmt.Println()
	fmt.Println("  Domain akan otomatis terdaftar saat pertama kali aplikasi client")
	fmt.Println("  melakukan validasi lisensi.")
	fmt.Println()
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func findOrCreatePlan(svc *license.Service, siteLimit int, duration, label string) (*license.Plan, error) {
	plans, err := svc.ListPlans()
	if err != nil {
		return nil, err
	}

	for _, p := range plans {
		if p.SiteLimit == siteLimit && p.Duration == duration && p.IsActive {
			return &p, nil
		}
	}

	slug := buildSlug(label, duration)
	return svc.CreatePlan(label+" "+durationLabel(duration), slug, siteLimit, duration, 0)
}

func buildSlug(label, duration string) string {
	s := strings.ToLower(label)
	s = strings.ReplaceAll(s, " ", "_")
	return s + "_" + duration
}

func durationLabel(d string) string {
	switch d {
	case "3y":
		return "3 Tahun"
	case "lifetime":
		return "Lifetime"
	}
	return "1 Tahun"
}

func sanitizeDomain(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "https://")
	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimSuffix(s, "/")
	return strings.ToLower(s)
}

func prompt(scanner *bufio.Scanner, label string) string {
	fmt.Printf("%s: ", label)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func promptChoice(scanner *bufio.Scanner, label string, valid []string) string {
	for {
		fmt.Printf("%s: ", label)
		scanner.Scan()
		val := strings.TrimSpace(scanner.Text())
		for _, v := range valid {
			if val == v {
				return val
			}
		}
		fmt.Printf("   ✗ Pilihan tidak valid. Masukkan salah satu dari: %s\n", strings.Join(valid, ", "))
	}
}

func printHeader() {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════╗")
	fmt.Println("  ║    ZyCrypt — Buat Lisensi Baru       ║")
	fmt.Println("  ╚══════════════════════════════════════╝")
	fmt.Println()
}

func printStep(n int, title string) {
	fmt.Println()
	fmt.Printf("  ── Step %d: %s ──\n", n, title)
}

func printDivider() {
	fmt.Println("  ──────────────────────────────────────")
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "\n  ✗ Error: %s\n\n", msg)
	os.Exit(1)
}
