package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	internalgit "github.com/h3y6e/skills/internal/git"
	"github.com/spf13/cobra"
)

const setupGitHubActionsOutputPath = ".github/workflows/skills-update.yml"

type setupGitHubActionsTemplateData struct {
	DestDir     string
	GitHubToken string
}

//go:embed templates/skills-update.yml.tmpl
var setupGitHubActionsTemplates embed.FS

func newSetupGitHubActionsCmd() *cobra.Command {
	var (
		destFlag  string
		forceFlag bool
	)

	c := &cobra.Command{
		Use:   "setup-github-actions",
		Short: "Generate a GitHub Actions workflow for automated skill updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupGitHubActions(cmd, destFlag, forceFlag)
		},
	}

	bindDestFlag(c, &destFlag, "Destination directory for installed skills")
	c.Flags().BoolVar(&forceFlag, "force", false, "Overwrite an existing workflow file")

	return c
}

func runSetupGitHubActions(cmd *cobra.Command, destDir string, force bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	if !internalgit.IsRepository(cmd.Context(), cwd) {
		fmt.Fprintln(cmd.ErrOrStderr(), "warning: current directory is not a git repository; generating workflow anyway")
	}

	outputPath := filepath.Join(cwd, setupGitHubActionsOutputPath)
	if !force {
		if _, err := os.Stat(outputPath); err == nil {
			return fmt.Errorf("%s already exists; re-run with --force to overwrite", outputPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check existing workflow: %w", err)
		}
	}

	rendered, err := renderSetupGitHubActionsTemplate(destDir)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create workflow directory: %w", err)
	}
	if err := os.WriteFile(outputPath, rendered, 0o644); err != nil {
		return fmt.Errorf("write workflow: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "generated %s\n", setupGitHubActionsOutputPath)
	return nil
}

func renderSetupGitHubActionsTemplate(destDir string) ([]byte, error) {
	tmpl, err := template.ParseFS(setupGitHubActionsTemplates, "templates/skills-update.yml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parse workflow template: %w", err)
	}

	var buf bytes.Buffer
	data := setupGitHubActionsTemplateData{
		DestDir:     destDir,
		GitHubToken: "${{ github.token }}",
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("render workflow template: %w", err)
	}

	return buf.Bytes(), nil
}
