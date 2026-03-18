package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	. "github.com/opd-ai/createon"
	"github.com/opd-ai/createon/pkg/files"
	"github.com/opd-ai/createon/pkg/subscription"
	"github.com/opd-ai/paywall"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "Manage subscriptions",
	Long:  `Commands for verifying and listing user subscriptions.`,
}

// subVerifyCmd verifies a user's access to a tier
var subVerifyCmd = &cobra.Command{
	Use:   "verify [email] [creator] [tier]",
	Short: "Verify a user's subscription access",
	Long: `Check if a user has access to a specific tier for a creator.

Example:
  createon sub verify user@example.com mycreator tier1`,
	Args: cobra.ExactArgs(3),
	RunE: runSubVerify,
}

// subListCmd lists subscriptions for a creator
var subListCmd = &cobra.Command{
	Use:   "list [creator]",
	Short: "List subscriptions for a creator",
	Long: `Display all active subscriptions for the specified creator.

Example:
  createon sub list mycreator`,
	Args: cobra.ExactArgs(1),
	RunE: runSubList,
}

func init() {
	rootCmd.AddCommand(subCmd)
	subCmd.AddCommand(subVerifyCmd)
	subCmd.AddCommand(subListCmd)
}

func runSubVerify(cmd *cobra.Command, args []string) error {
	email := args[0]
	creatorUsername := args[1]
	tierName := args[2]

	// Load config
	configPath := viper.GetString("config")
	if configPath == "" {
		configPath = "config/server.yaml"
	}
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create file manager
	fm, err := files.NewManager("data")
	if err != nil {
		return fmt.Errorf("failed to create file manager: %w", err)
	}

	// Create paywall (nil is acceptable for verify-only operations)
	var pw *paywall.Paywall

	// Create subscription manager
	mgr := subscription.NewManager(fm, pw, "data", cfg)

	// Verify access
	ctx := context.Background()
	hasAccess := mgr.VerifyAccess(ctx, email, creatorUsername, tierName)

	if hasAccess {
		fmt.Println("Access: granted")
	} else {
		fmt.Println("Access: denied")
	}

	return nil
}

func runSubList(cmd *cobra.Command, args []string) error {
	creatorUsername := args[0]

	// Read subscriptions from data directory
	subDir := filepath.Join("data", "creators", creatorUsername, "subscriptions")
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		fmt.Println("No active subscriptions")
		return nil
	}

	entries, err := os.ReadDir(subDir)
	if err != nil {
		return fmt.Errorf("failed to read subscriptions: %w", err)
	}

	// Filter and collect active subscriptions
	var subs []Subscription
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		subPath := filepath.Join(subDir, entry.Name())
		data, err := os.ReadFile(subPath)
		if err != nil {
			continue
		}

		var sub Subscription
		if err := yaml.Unmarshal(data, &sub); err != nil {
			continue
		}

		// Only include subscriptions that haven't expired
		if time.Now().Before(sub.ExpiresAt) {
			subs = append(subs, sub)
		}
	}

	if len(subs) == 0 {
		fmt.Println("No active subscriptions")
		return nil
	}

	// Print subscriptions table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "USER\tTIER\tEXPIRES")
	fmt.Fprintln(w, "----\t----\t-------")
	for _, sub := range subs {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			sub.Email,
			sub.TierID,
			sub.ExpiresAt.Format("2006-01-02"),
		)
	}
	w.Flush()

	return nil
}
