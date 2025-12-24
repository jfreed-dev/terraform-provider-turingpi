# Contributing to Terraform Provider for Turing Pi

Thank you for your interest in contributing to this project!

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/terraform-provider-turingpi.git`
3. Create a branch: `git checkout -b feature/your-feature`

## Development Setup

```bash
# Install dependencies
go mod tidy

# Build
go build -o terraform-provider-turingpi

# Run tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...
```

## Making Changes

1. Write clear, concise commit messages
2. Add tests for new functionality
3. Ensure all tests pass before submitting
4. Update documentation if needed

## Code Style

- Follow standard Go conventions (`gofmt`, `golint`)
- Run `golangci-lint run` before submitting
- Keep functions focused and well-documented

## Pull Request Process

1. Update the CHANGELOG.md with your changes
2. Ensure CI checks pass
3. Request review from maintainers
4. Address any feedback

## Release Process (Maintainers)

Releases are automated via Makefile targets. All releases are GPG-signed and published to the [Terraform Registry](https://registry.terraform.io/providers/jfreed-dev/turingpi).

### Creating a Release

```bash
# Full release: updates docs, runs tests, commits, tags, and pushes
make release VERSION=1.0.10
```

This command will:
1. Update version references in `README.md`, `docs/`, and `examples/`
2. Run the full test suite
3. Commit the version updates (GPG-signed)
4. Create a signed tag `v1.0.10`
5. Push to origin (triggers GitHub Actions release workflow)

### Partial Release Prep

```bash
# Just update version numbers (no commit/tag)
make release-prep VERSION=1.0.10
```

### Post-Release

After the release workflow completes:
1. Verify the release on [GitHub Releases](https://github.com/jfreed-dev/terraform-provider-turingpi/releases)
2. Confirm it appears on [Terraform Registry](https://registry.terraform.io/providers/jfreed-dev/turingpi)
3. Update `CHANGELOG.md` with release notes

### Requirements

- GPG key configured for commit/tag signing (`git config commit.gpgsign true`)
- GPG key registered on GitHub for verified badges
- Push access to the repository

## Testing Against Real Hardware

If you have access to a Turing Pi 2.5:

1. Set environment variables:
   ```bash
   export TURINGPI_USERNAME=root
   export TURINGPI_PASSWORD=turing
   export TURINGPI_ENDPOINT=https://your-turingpi-ip
   ```

2. Create a test Terraform configuration
3. Run `terraform plan` and `terraform apply`

## Reporting Issues

- Use the issue templates provided
- Include Terraform version, provider version, and Go version
- Provide minimal reproduction steps

## Questions?

Open a discussion or issue if you have questions about contributing.
