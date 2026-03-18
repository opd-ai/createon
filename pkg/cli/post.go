package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// postCmd represents the post command
var postCmd = &cobra.Command{
	Use:   "post",
	Short: "Manage creator posts",
	Long:  `Commands for publishing and managing creator content posts.`,
}

// postPublishCmd represents the post publish command
var postPublishCmd = &cobra.Command{
	Use:   "publish [username] [file.md]",
	Short: "Publish a new post for a creator",
	Long: `Publish a markdown file as a new post for the specified creator.

Example:
  createon post publish mycreator content.md -t "My Post Title" -r tier1`,
	Args: cobra.ExactArgs(2),
	RunE: runPostPublish,
}

// PostMetadata holds the metadata for a published post
type PostMetadata struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title"`
	Tier      string    `yaml:"tier"`
	Tags      []string  `yaml:"tags,omitempty"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

var (
	postTitle string
	postTier  string
	postTags  string
)

func init() {
	rootCmd.AddCommand(postCmd)
	postCmd.AddCommand(postPublishCmd)

	postPublishCmd.Flags().StringVarP(&postTitle, "title", "t", "", "Post title (required)")
	postPublishCmd.Flags().StringVarP(&postTier, "tier", "r", "free", "Required tier for access")
	postPublishCmd.Flags().StringVar(&postTags, "tags", "", "Comma-separated tags")
	postPublishCmd.MarkFlagRequired("title")
}

func runPostPublish(cmd *cobra.Command, args []string) error {
	username := args[0]
	inputFile := args[1]

	// Read markdown content
	content, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Generate post ID
	postID := uuid.New().String()

	// Create post directory
	postDir := filepath.Join("data", "creators", username, "posts")
	if err := os.MkdirAll(postDir, 0o755); err != nil {
		return fmt.Errorf("failed to create post directory: %w", err)
	}

	// Write markdown content
	postFile := filepath.Join(postDir, postID+".md")
	if err := os.WriteFile(postFile, content, 0o644); err != nil {
		return fmt.Errorf("failed to write post content: %w", err)
	}

	// Create metadata
	now := time.Now()
	meta := PostMetadata{
		ID:        postID,
		Title:     postTitle,
		Tier:      postTier,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Parse tags if provided
	if postTags != "" {
		meta.Tags = parseTags(postTags)
	}

	// Write metadata
	metaFile := filepath.Join(postDir, postID+".yaml")
	metaData, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaFile, metaData, 0o644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	fmt.Printf("Published post: %s\n", postID)
	fmt.Printf("  Title: %s\n", postTitle)
	fmt.Printf("  Tier: %s\n", postTier)
	fmt.Printf("  File: %s\n", postFile)

	return nil
}

// parseTags splits comma-separated tags and trims whitespace
func parseTags(tags string) []string {
	parts := strings.Split(tags, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
