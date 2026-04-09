package cmd

import (
	"fmt"
	"os"

	"github.com/h3y6e/skills/internal/git"
	"github.com/h3y6e/skills/internal/lock"
	"github.com/h3y6e/skills/internal/skill"
	"github.com/spf13/cobra"
)

func runAdd(cmd *cobra.Command, source string, listMode bool, skillFilter []string, destDir string) error {
	layout := lock.NewLayout(destDir)

	ref, err := skill.ParseSource(source)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "skills-add-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := git.ShallowClone(cmd.Context(), ref.CloneURL, ref.Ref, tmpDir); err != nil {
		return fmt.Errorf("clone %s: %w", ref.CanonicalSource, err)
	}

	discovered, err := skill.DiscoverSkills(tmpDir)
	if err != nil {
		return fmt.Errorf("discover skills: %w", err)
	}

	discovered = skill.FilterDiscoveredSkills(discovered, skillFilter)

	if listMode {
		for _, s := range discovered {
			fmt.Fprintln(cmd.OutOrStdout(), s.Name)
		}
		return nil
	}

	if len(discovered) == 0 {
		return fmt.Errorf("no skills found to install")
	}

	return skill.InstallSkills(discovered, ref, layout)
}
