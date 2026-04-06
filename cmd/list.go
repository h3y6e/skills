package cmd

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/h3y6e/skills/internal/lock"
	"github.com/spf13/cobra"
)

type listEntryJSON struct {
	SkillName    string `json:"skillName"`
	Source       string `json:"source"`
	SourceType   string `json:"sourceType"`
	ComputedHash string `json:"computedHash"`
}

func runList(cmd *cobra.Command, destDir string, jsonFlag bool) error {
	layout := lock.NewLayout(destDir)
	lf, err := loadLockFile(layout)
	if err != nil {
		return err
	}
	entries := lock.FilterEntriesByDest(lf.Skills, layout.DestDir)

	names := slices.Sorted(maps.Keys(entries))

	if jsonFlag {
		payload := make([]listEntryJSON, 0, len(names))
		for _, name := range names {
			entry := entries[name]
			payload = append(payload, listEntryJSON{
				SkillName:    name,
				Source:       entry.Source,
				SourceType:   entry.SourceType,
				ComputedHash: entry.ComputedHash,
			})
		}
		data, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return err
	}

	if len(names) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "no installed skills")
		return nil
	}

	for _, name := range names {
		entry := entries[name]
		fmt.Fprintf(cmd.OutOrStdout(), "%s  %s  %s\n", name, entry.Source, ShortHash(entry.ComputedHash))
	}
	return nil
}
