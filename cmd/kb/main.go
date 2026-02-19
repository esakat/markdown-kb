package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/esakat/markdown-kb/internal/config"
	"github.com/esakat/markdown-kb/internal/index"
	"github.com/esakat/markdown-kb/internal/scanner"
	"github.com/esakat/markdown-kb/internal/server"
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
		os.Exit(1)
	}
}

func validateRootDir(rootDir string) error {
	info, err := os.Stat(rootDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory %q does not exist", rootDir)
	}
	if err != nil {
		return fmt.Errorf("cannot read directory %q: %w", rootDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q is not a directory", rootDir)
	}
	return nil
}

func scanAndIndex(rootDir string) (*index.Store, []scanner.Document, error) {
	docs, err := scanner.Scan(rootDir)
	if err != nil {
		return nil, nil, fmt.Errorf("scanning directory: %w", err)
	}

	if len(docs) == 0 {
		fmt.Fprintf(os.Stderr, "Warning: no markdown files found in %q\n", rootDir)
	}

	store, err := index.New()
	if err != nil {
		return nil, nil, fmt.Errorf("creating index: %w", err)
	}

	for _, doc := range docs {
		if err := store.IndexDocument(doc); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to index %q: %v\n", doc.RelPath, err)
		}
	}

	return store, docs, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
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

			if err := validateRootDir(cfg.RootDir); err != nil {
				return err
			}

			store, docs, err := scanAndIndex(cfg.RootDir)
			if err != nil {
				return err
			}
			defer store.Close()

			srv := server.New(cfg, store)

			// Graceful shutdown
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 1)
			go func() {
				errCh <- srv.Start()
			}()

			url := fmt.Sprintf("http://localhost:%d", cfg.Port)
			fmt.Printf("Serving %d documents from %s on :%d\n", len(docs), cfg.RootDir, cfg.Port)

			if cfg.Open {
				openBrowser(url)
			}

			select {
			case <-ctx.Done():
				fmt.Println("\nShutting down...")
				return srv.Shutdown(context.Background())
			case err := <-errCh:
				if err != nil && err != http.ErrServerClosed {
					return fmt.Errorf("server error: %w", err)
				}
				return nil
			}
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

			if err := validateRootDir(rootDir); err != nil {
				return err
			}

			docs, err := scanner.Scan(rootDir)
			if err != nil {
				return fmt.Errorf("scanning directory: %w", err)
			}

			if len(docs) == 0 {
				fmt.Fprintf(os.Stderr, "Warning: no markdown files found in %q\n", rootDir)
			}

			switch format {
			case "json":
				return outputJSON(docs)
			case "text":
				return outputText(docs)
			default:
				return fmt.Errorf("unknown format %q (use json or text)", format)
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "json", "Output format (json|text)")

	return cmd
}

type indexEntry struct {
	Path   string         `json:"path"`
	Title  string         `json:"title"`
	Status string         `json:"status,omitempty"`
	Tags   []string       `json:"tags,omitempty"`
	Size   int64          `json:"size"`
	Meta   map[string]any `json:"meta,omitempty"`
}

func docToEntry(doc scanner.Document) indexEntry {
	entry := indexEntry{
		Path: doc.RelPath,
		Size: doc.Size,
		Meta: doc.Frontmatter,
	}
	if doc.Frontmatter != nil {
		entry.Title, _ = doc.Frontmatter["title"].(string)
		entry.Status, _ = doc.Frontmatter["status"].(string)
		if tags, ok := doc.Frontmatter["tags"].([]any); ok {
			for _, t := range tags {
				if s, ok := t.(string); ok {
					entry.Tags = append(entry.Tags, s)
				}
			}
		}
	}
	return entry
}

func outputJSON(docs []scanner.Document) error {
	entries := make([]indexEntry, len(docs))
	for i, doc := range docs {
		entries[i] = docToEntry(doc)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(entries)
}

func outputText(docs []scanner.Document) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PATH\tTITLE\tSTATUS\tTAGS")
	fmt.Fprintln(w, "----\t-----\t------\t----")

	for _, doc := range docs {
		entry := docToEntry(doc)
		tags := strings.Join(entry.Tags, ", ")
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", entry.Path, entry.Title, entry.Status, tags)
	}

	return w.Flush()
}
