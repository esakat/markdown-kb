package main

import (
	"fmt"
	"os"

	"github.com/esakat/markdown-kb/internal/config"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:     "kb",
		Short:   "Markdown Knowledge Base viewer for Git repositories",
		Version: version,
	}

	rootCmd.AddCommand(newServeCmd())
	rootCmd.AddCommand(newIndexCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newServeCmd() *cobra.Command {
	var cfg config.ServeConfig

	cmd := &cobra.Command{
		Use:   "serve [path]",
		Short: "Start the web server",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				cfg.RootDir = args[0]
			}
			if cfg.RootDir == "" {
				wd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("getting working directory: %w", err)
				}
				cfg.RootDir = wd
			}
			fmt.Printf("Starting server on :%d for %s\n", cfg.Port, cfg.RootDir)
			// TODO: implement server start
			return nil
		},
	}

	cmd.Flags().IntVar(&cfg.Port, "port", 3000, "Port to listen on")
	cmd.Flags().BoolVar(&cfg.Open, "open", false, "Open browser after starting")

	return cmd
}

func newIndexCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "index [path]",
		Short: "Build search index and output metadata",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rootDir := "."
			if len(args) > 0 {
				rootDir = args[0]
			}
			fmt.Printf("Indexing %s (format: %s)\n", rootDir, format)
			// TODO: implement indexing
			return nil
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Output format (json|text)")

	return cmd
}
