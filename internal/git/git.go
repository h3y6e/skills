package git

import (
	"context"
	"fmt"
	"os/exec"
)

func gitCommand(ctx context.Context, args ...string) (*exec.Cmd, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found: %w", err)
	}
	return exec.CommandContext(ctx, gitPath, args...), nil
}

// ShallowClone clones cloneURL into destDir with depth=1.
func ShallowClone(ctx context.Context, cloneURL, destDir string) error {
	cmd, err := gitCommand(ctx, "clone", "--depth", "1", cloneURL, destDir)
	if err != nil {
		return err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone %q: %w\n%s", cloneURL, err, out)
	}
	return nil
}

// DiffNoIndex runs `git diff --no-index` between two paths.
// Returns the diff output, whether a diff exists, and any error.
// Exit code 1 indicates a diff (not an error).
func DiffNoIndex(ctx context.Context, pathA, pathB string) (string, bool, error) {
	cmd, err := gitCommand(ctx, "--no-pager", "diff", "--no-index", "--", pathA, pathB)
	if err != nil {
		return "", false, err
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return string(out), true, nil
		}
		return "", false, fmt.Errorf("git diff --no-index: %w\n%s", err, out)
	}
	return string(out), false, nil
}
