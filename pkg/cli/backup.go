// Package cli provides command-line interface commands for the Createon platform.
package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup and restore data",
	Long:  `Commands for creating and restoring backups of the data directory.`,
}

// backupCreateCmd creates a backup archive
var backupCreateCmd = &cobra.Command{
	Use:   "create [file.tar.gz]",
	Short: "Create a backup archive",
	Long: `Create a compressed tar archive of the data directory.

Example:
  createon backup create /tmp/backup.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupCreate,
}

// backupRestoreCmd restores from a backup archive
var backupRestoreCmd = &cobra.Command{
	Use:   "restore [file.tar.gz]",
	Short: "Restore from a backup archive",
	Long: `Extract a backup archive to restore the data directory.

Example:
  createon backup restore /tmp/backup.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: runBackupRestore,
}

var forceRestore bool

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)

	backupRestoreCmd.Flags().BoolVarP(&forceRestore, "force", "f", false, "Force restore without confirmation")
}

func runBackupCreate(cmd *cobra.Command, args []string) error {
	outputFile := args[0]
	dataDir := "data"

	// Check data directory exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		return fmt.Errorf("data directory does not exist: %s", dataDir)
	}

	// Create output file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	// Create gzip writer
	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	// Create tar writer
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// Walk data directory and add files
	err = filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return addToTar(tw, path, info)
	})
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	fmt.Printf("Backup created: %s\n", outputFile)
	return nil
}

func addToTar(tw *tar.Writer, path string, info os.FileInfo) error {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = path

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func runBackupRestore(cmd *cobra.Command, args []string) error {
	inputFile := args[0]

	// Check if data directory exists and warn
	if _, err := os.Stat("data"); err == nil && !forceRestore {
		fmt.Print("Warning: data directory exists and will be overwritten. Continue? [y/N] ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Restore cancelled")
			return nil
		}
	}

	// Open backup file
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to read gzip: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Extract files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar: %w", err)
		}

		if err := extractFromTar(tr, header); err != nil {
			return fmt.Errorf("failed to extract %s: %w", header.Name, err)
		}
	}

	fmt.Printf("Backup restored from: %s\n", inputFile)
	return nil
}

func extractFromTar(tr *tar.Reader, header *tar.Header) error {
	target := header.Name

	switch header.Typeflag {
	case tar.TypeDir:
		return os.MkdirAll(target, os.FileMode(header.Mode))
	case tar.TypeReg:
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tr)
		return err
	default:
		return nil
	}
}
