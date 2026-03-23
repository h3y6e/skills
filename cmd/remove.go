package cmd

import (
	"fmt"
	"os"

	"github.com/h3y6e/skills/internal/lock"
	"github.com/spf13/cobra"
)

func runRemove(cmd *cobra.Command, skillNames []string, destDir string) error {
	layout := lock.NewLayout(destDir)
	lf, err := loadLockFile(layout)
	if err != nil {
		return err
	}

	for _, name := range skillNames {
		if _, ok := lf.Skills[name]; !ok {
			return fmt.Errorf("skill %q not found in lockfile", name)
		}
	}

	for _, name := range skillNames {
		skillDir := layout.SkillDir(name)
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
