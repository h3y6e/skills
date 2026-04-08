package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/h3y6e/skills/internal/lock"
	"github.com/spf13/cobra"
)

func runRemove(cmd *cobra.Command, skillNames []string, destDir string) error {
	layout := lock.NewLayout(destDir)
	lf, err := loadLockFile(layout)
	if err != nil {
		return err
	}
	entries := lock.FilterEntriesByDest(lf.Skills, layout.DestDir)

	for _, name := range skillNames {
		if _, ok := entries[name]; !ok {
			return fmt.Errorf("skill %q not found in lockfile for dest %q", name, layout.DestDir)
		}
	}

	for _, name := range skillNames {
		skillDir := filepath.Join(lock.EffectiveDest(entries[name]), name)
		if err := os.RemoveAll(skillDir); err != nil {
			return fmt.Errorf("remove skill directory %q: %w", name, err)
		}

		delete(lf.Skills, name)

		fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", name)
	}

	if err := lock.WriteFile(layout.LockPath(), lf); err != nil {
		return fmt.Errorf("write lockfile: %w", err)
	}

	return nil
}
