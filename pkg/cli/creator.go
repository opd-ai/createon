// pkg/cli/creator.go
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/opd-ai/createon"
	"github.com/opd-ai/createon/pkg/files"
)

func init() {
	creatorCmd := &cobra.Command{
		Use:   "creator",
		Short: "Manage creators",
	}

	addCmd := &cobra.Command{
		Use:   "add [username]",
		Short: "Add new creator",
		Args:  cobra.ExactArgs(1),
		RunE:  runAddCreator,
	}
	addCmd.Flags().StringP("name", "n", "", "display name")
	addCmd.Flags().StringP("bio", "b", "", "creator bio")
	addCmd.Flags().StringSliceP("tiers", "t", []string{}, "subscription tiers (name:price_btc:price_xmr)")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List creators",
		RunE:  runListCreators,
	}

	creatorCmd.AddCommand(addCmd, listCmd)
	rootCmd.AddCommand(creatorCmd)
}

func runAddCreator(cmd *cobra.Command, args []string) error {
	username := args[0]
	name, _ := cmd.Flags().GetString("name")
	bio, _ := cmd.Flags().GetString("bio")
	tiers, _ := cmd.Flags().GetStringSlice("tiers")

	fm, err := files.NewManager("data")
	if err != nil {
		return fmt.Errorf("File manager creation error: %w")
	}

	// Create creator directory
	creatorDir := filepath.Join("data", "creators", username)
	if err := os.MkdirAll(creatorDir, 0755); err != nil {
		return fmt.Errorf("failed to create creator directory: %w", err)
	}

	// Parse tiers
	var tierConfigs []createon.Tier
	for i, t := range tiers {
		var name, btcStr, xmrStr string
		fmt.Sscanf(t, "%s:%s:%s", &name, &btcStr, &xmrStr)

		// Convert prices to float64
		btc, err := strconv.ParseFloat(btcStr, 64)
		if err != nil {
			return fmt.Errorf("invalid BTC price for tier %s: %w", name, err)
		}

		xmr, err := strconv.ParseFloat(xmrStr, 64)
		if err != nil {
			return fmt.Errorf("invalid XMR price for tier %s: %w", name, err)
		}

		tierConfigs = append(tierConfigs, createon.Tier{
			ID:       fmt.Sprintf("tier%d", i+1),
			Name:     name,
			PriceBTC: btc,
			PriceXMR: xmr,
		})
	}

	// Create creator config
	creator := createon.Creator{
		Username:    username,
		DisplayName: name,
		Bio:         bio,
		Tiers:       tierConfigs,
	}

	// Save config
	configPath := filepath.Join(creatorDir, "config.yaml")
	return fm.WriteYAML(configPath, creator)
}

func runListCreators(cmd *cobra.Command, args []string) error {
	fm, err := files.NewManager("data")
	if err != nil {
		return fmt.Errorf("File manager creation error: %w")
	}
	creatorsDir := filepath.Join("data", "creators")

	// Read creator directories
	entries, err := os.ReadDir(creatorsDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No creators found")
			return nil
		}
		return fmt.Errorf("failed to list creators: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var creator createon.Creator
		configPath := filepath.Join(creatorsDir, entry.Name(), "config.yaml")
		if err := fm.ReadYAML(configPath, &creator); err != nil {
			continue
		}

		fmt.Printf("%s (%s)\n", creator.Username, creator.DisplayName)
		fmt.Printf("Bio: %s\n", creator.Bio)
		fmt.Printf("Tiers: %d\n", len(creator.Tiers))
		for _, tier := range creator.Tiers {
			fmt.Printf("  - %s: %.8f BTC / %.8f XMR\n",
				tier.Name, tier.PriceBTC, tier.PriceXMR)
		}
		fmt.Println("---")
	}

	return nil
}
