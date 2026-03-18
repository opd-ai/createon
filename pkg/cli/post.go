package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// postHistoryCmd shows version history of a post
var postHistoryCmd = &cobra.Command{
	Use:   "history [username] [post-id]",
	Short: "Show version history of a post",
	Long: `Display all versions of a post with timestamps and sizes.

Example:
  createon post history mycreator abc-123-def`,
	Args: cobra.ExactArgs(2),
	RunE: runPostHistory,
}

// postRevertCmd reverts a post to a previous version
var postRevertCmd = &cobra.Command{
	Use:   "revert [username] [post-id] [version]",
	Short: "Revert a post to a previous version",
	Long: `Restore a post to a specific version number.

Example:
  createon post revert mycreator abc-123-def 1`,
	Args: cobra.ExactArgs(3),
	RunE: runPostRevert,
}

// PostMetadata holds the metadata for a published post
type PostMetadata struct {
	ID        string    `yaml:"id"`
	Title     string    `yaml:"title"`
	Tier      string    `yaml:"tier"`
	Tags      []string  `yaml:"tags,omitempty"`
	Version   int       `yaml:"version"`
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
	postCmd.AddCommand(postPublishCmd, postHistoryCmd, postRevertCmd)

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

	// Create versioned post directory
	postDir := filepath.Join("data", "creators", username, "posts", postID)
	if err := os.MkdirAll(postDir, 0o755); err != nil {
		return fmt.Errorf("failed to create post directory: %w", err)
	}

	// Write markdown content as version 1
	postFile := filepath.Join(postDir, "v1.md")
	if err := os.WriteFile(postFile, content, 0o644); err != nil {
		return fmt.Errorf("failed to write post content: %w", err)
	}

	// Create metadata with version
	now := time.Now()
	meta := PostMetadata{
		ID:        postID,
		Title:     postTitle,
		Tier:      postTier,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Parse tags if provided
	if postTags != "" {
		meta.Tags = parseTags(postTags)
	}

	// Write metadata
	metaFile := filepath.Join(postDir, "metadata.yaml")
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
	fmt.Printf("  Version: 1\n")
	fmt.Printf("  File: %s\n", postFile)

	return nil
}

// runPostHistory displays version history for a post
func runPostHistory(cmd *cobra.Command, args []string) error {
	username := args[0]
	postID := args[1]

	postDir := filepath.Join("data", "creators", username, "posts", postID)

	// Read metadata
	metaFile := filepath.Join(postDir, "metadata.yaml")
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	var meta PostMetadata
	if err := yaml.Unmarshal(metaData, &meta); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// List all version files
	versions, err := listVersions(postDir)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	fmt.Printf("Post: %s\n", meta.Title)
	fmt.Printf("ID: %s\n", postID)
	fmt.Printf("Current version: %d\n", meta.Version)
	fmt.Printf("\nVersion History:\n")

	for _, v := range versions {
		versionFile := filepath.Join(postDir, fmt.Sprintf("v%d.md", v))
		info, err := os.Stat(versionFile)
		if err != nil {
			continue
		}
		current := ""
		if v == meta.Version {
			current = " (current)"
		}
		fmt.Printf("  v%d: %s  %d bytes%s\n",
			v, info.ModTime().Format(time.RFC3339), info.Size(), current)
	}

	return nil
}

// runPostRevert restores a post to a previous version
func runPostRevert(cmd *cobra.Command, args []string) error {
	username := args[0]
	postID := args[1]
	targetVersion, err := strconv.Atoi(args[2])
	if err != nil {
		return fmt.Errorf("invalid version number: %w", err)
	}

	postDir := filepath.Join("data", "creators", username, "posts", postID)

	// Read current metadata
	metaFile := filepath.Join(postDir, "metadata.yaml")
	metaData, err := os.ReadFile(metaFile)
	if err != nil {
		return fmt.Errorf("post not found: %w", err)
	}

	var meta PostMetadata
	if err := yaml.Unmarshal(metaData, &meta); err != nil {
		return fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Check target version exists
	targetFile := filepath.Join(postDir, fmt.Sprintf("v%d.md", targetVersion))
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return fmt.Errorf("version %d not found", targetVersion)
	}

	// Read target version content
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	// Create new version from the reverted content
	newVersion := meta.Version + 1
	newFile := filepath.Join(postDir, fmt.Sprintf("v%d.md", newVersion))
	if err := os.WriteFile(newFile, content, 0o644); err != nil {
		return fmt.Errorf("failed to write reverted content: %w", err)
	}

	// Update metadata
	meta.Version = newVersion
	meta.UpdatedAt = time.Now()

	newMetaData, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metaFile, newMetaData, 0o644); err != nil {
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	fmt.Printf("Reverted post to version %d\n", targetVersion)
	fmt.Printf("  New version: %d\n", newVersion)
	fmt.Printf("  Title: %s\n", meta.Title)

	return nil
}

// listVersions returns all version numbers for a post in ascending order
func listVersions(postDir string) ([]int, error) {
	entries, err := os.ReadDir(postDir)
	if err != nil {
		return nil, err
	}

	var versions []int
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "v") && strings.HasSuffix(name, ".md") {
			vStr := name[1 : len(name)-3]
			if v, err := strconv.Atoi(vStr); err == nil {
				versions = append(versions, v)
			}
		}
	}

	sort.Ints(versions)
	return versions, nil
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
