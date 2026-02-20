package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Commit represents a simplified Git commit.
type Commit struct {
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}

// BlameLine represents a single line's blame info.
type BlameLine struct {
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	LineNo  int       `json:"line_no"`
	Content string    `json:"content"`
}

// FileHistory returns the commit history for a specific file (most recent first).
func FileHistory(repoDir, filePath string) ([]Commit, error) {
	cmd := exec.Command("git", "log", "--pretty=format:%H\x1f%an\x1f%aI\x1f%s", "--follow", "--", filePath)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			if strings.Contains(stderr, "not a git repository") {
				return nil, fmt.Errorf("not a git repository: %s", repoDir)
			}
		}
		return nil, fmt.Errorf("git log: %w", err)
	}

	output := strings.TrimSpace(string(out))
	if output == "" {
		return []Commit{}, nil
	}

	var commits []Commit
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, "\x1f", 4)
		if len(parts) != 4 {
			continue
		}
		date, _ := time.Parse(time.RFC3339, parts[2])
		commits = append(commits, Commit{
			Hash:    parts[0],
			Author:  parts[1],
			Date:    date,
			Message: parts[3],
		})
	}

	return commits, nil
}

// Diff returns the unified diff between two commits for a file.
func Diff(repoDir, filePath, fromHash, toHash string) (string, error) {
	cmd := exec.Command("git", "diff", fromHash+".."+toHash, "--", filePath)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// git diff exits 1 when there are differences (normal behavior)
			if exitErr.ExitCode() == 1 {
				return string(out), nil
			}
		}
		return "", fmt.Errorf("git diff: %w", err)
	}

	return string(out), nil
}

// Blame returns line-by-line attribution for the entire file.
func Blame(repoDir, filePath string) ([]BlameLine, error) {
	return blameExec(repoDir, filePath, nil)
}

// BlameRange returns blame information for a specific line range [startLine, endLine].
func BlameRange(repoDir, filePath string, startLine, endLine int) ([]BlameLine, error) {
	lineRange := fmt.Sprintf("%d,%d", startLine, endLine)
	return blameExec(repoDir, filePath, &lineRange)
}

func blameExec(repoDir, filePath string, lineRange *string) ([]BlameLine, error) {
	args := []string{"blame", "--porcelain"}
	if lineRange != nil {
		args = append(args, "-L", *lineRange)
	}
	args = append(args, "--", filePath)

	cmd := exec.Command("git", args...)
	cmd.Dir = repoDir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git blame: %w", err)
	}

	return parsePorcelainBlame(string(out))
}

type commitInfo struct {
	Author string
	Date   time.Time
}

func parsePorcelainBlame(output string) ([]BlameLine, error) {
	var lines []BlameLine
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Cache commit info since porcelain only shows details on first occurrence
	cache := make(map[string]*commitInfo)

	var current BlameLine
	var currentInfo *commitInfo

	for scanner.Scan() {
		line := scanner.Text()

		// Commit header line: <40-char hash> <orig_line> <final_line> [<num_lines>]
		if len(line) >= 40 && !strings.HasPrefix(line, "\t") {
			parts := strings.Fields(line)
			if len(parts) >= 3 && len(parts[0]) == 40 {
				lineNo, _ := strconv.Atoi(parts[2])
				hash := parts[0]
				current = BlameLine{
					Hash:   hash,
					LineNo: lineNo,
				}

				if cached, ok := cache[hash]; ok {
					currentInfo = cached
				} else {
					currentInfo = &commitInfo{}
					cache[hash] = currentInfo
				}
				continue
			}
		}

		if strings.HasPrefix(line, "author ") {
			currentInfo.Author = strings.TrimPrefix(line, "author ")
		} else if strings.HasPrefix(line, "author-time ") {
			ts, err := strconv.ParseInt(strings.TrimPrefix(line, "author-time "), 10, 64)
			if err == nil {
				currentInfo.Date = time.Unix(ts, 0)
			}
		} else if strings.HasPrefix(line, "\t") {
			// Content line marks end of this blame entry
			current.Content = strings.TrimPrefix(line, "\t")
			current.Author = currentInfo.Author
			current.Date = currentInfo.Date
			lines = append(lines, current)
		}
	}

	return lines, scanner.Err()
}

// FileDates returns the first (created) and last (updated) commit dates for a file.
func FileDates(repoDir, filePath string) (created, updated time.Time, err error) {
	// Get the first commit date (oldest)
	cmdFirst := exec.Command("git", "log", "--diff-filter=A", "--follow", "--format=%aI", "--", filePath)
	cmdFirst.Dir = repoDir
	outFirst, err := cmdFirst.Output()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("git log (first): %w", err)
	}

	// Get the last commit date (newest)
	cmdLast := exec.Command("git", "log", "-1", "--format=%aI", "--", filePath)
	cmdLast.Dir = repoDir
	outLast, err := cmdLast.Output()
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("git log (last): %w", err)
	}

	firstDates := strings.TrimSpace(string(outFirst))
	if firstDates == "" {
		return time.Time{}, time.Time{}, nil
	}

	// First commit is the last line of the first command output
	dateLines := strings.Split(firstDates, "\n")
	createdStr := dateLines[len(dateLines)-1]
	created, _ = time.Parse(time.RFC3339, strings.TrimSpace(createdStr))

	lastDate := strings.TrimSpace(string(outLast))
	updated, _ = time.Parse(time.RFC3339, lastDate)

	return created, updated, nil
}
