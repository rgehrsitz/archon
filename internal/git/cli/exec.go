package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/rgehrsitz/archon/internal/errors"
	"github.com/rgehrsitz/archon/internal/logging"
)

// Config holds configuration for CLI Git operations
type Config struct {
	Path    string // Repository path
	GitPath string // Path to git executable (if not in PATH)
}

// Git execution helpers for Repository - the Repository struct is defined in repository.go

// exec executes a git command with the given arguments
func (r *Repository) exec(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.gitCmd, args...)
	cmd.Dir = r.config.Path

	// Set up environment
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0", // Disable interactive prompts
		"GIT_ASKPASS=true",      // Use credential helpers only
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	// Log the command execution
	logger := logging.Log()

	if err != nil {
		logger.Error().
			Err(err).
			Str("cmd", fmt.Sprintf("git %s", strings.Join(args, " "))).
			Str("dir", r.config.Path).
			Dur("duration", duration).
			Str("stderr", strings.TrimSpace(stderr.String())).
			Msg("Git command failed")
		return nil, fmt.Errorf("git %s: %w\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	logger.Debug().
		Str("cmd", fmt.Sprintf("git %s", strings.Join(args, " "))).
		Str("dir", r.config.Path).
		Dur("duration", duration).
		Msg("Git command completed successfully")
	return stdout.Bytes(), nil
}

// execLines executes a git command and returns output as lines
func (r *Repository) execLines(ctx context.Context, args ...string) ([]string, error) {
	output, err := r.exec(ctx, args...)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

// (ensureRepo previously enforced repository path; not used currently.)

// Helper to convert git CLI error to Archon error envelope
func (r *Repository) wrapError(err error, code string, message string) errors.Envelope {
	if err == nil {
		return errors.Envelope{}
	}
	return errors.WrapError(code, message, err)
}

// No Close method here - it's in repository.go
