# skills

A CLI tool to manage local [agent skills](https://agentskills.io), inspired by [vercel-labs/skills](https://github.com/vercel-labs/skills).

## Why

- Version-control skills from external repositories, say in your dotfiles
- Selectively install only the skills you need
- Review upstream diffs before applying changes

## Install

```sh
go install github.com/h3y6e/skills@latest
```

Or download a binary from [Releases](https://github.com/h3y6e/skills/releases).

## Usage

```sh
skills add h3y6e/spec-skills                               # Install skills from a repository
skills add h3y6e/spec-skills#main                          # Install from an explicit ref
skills add --list h3y6e/spec-skills                        # Preview available skills
skills add -s spec-plan h3y6e/spec-skills                  # Install specific skills only
skills add -d dot_agents/exact_skills h3y6e/spec-skills    # Custom destination (e.g. chezmoi)

skills list                                                # Show installed skills
skills check                                               # Check for upstream updates
skills update                                              # Review diffs and apply updates
skills update h3y6e/spec-skills#main                      # Update only entries installed from that ref
skills update -y                                           # Apply without confirmation (CI-safe)
skills remove spec-plan                                    # Remove a skill
```

### Directory structure

Skills are installed to `.agents/skills/` by default (configurable with `-d`). The lockfile format is compatible with [vercel-labs/skills](https://github.com/vercel-labs/skills).

```
.agents/skills/
├── spec-plan/
│   └── SKILL.md
└── spec-specify/
    └── SKILL.md
skills-lock.json
```

## License

[MIT](LICENSE)
