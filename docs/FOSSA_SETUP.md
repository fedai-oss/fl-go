# FOSSA Setup Guide

This guide explains how to set up and configure FOSSA for dependency analysis and compliance in the FL-GO project.

## What is FOSSA?

[FOSSA](https://fossa.com/) is a comprehensive dependency analysis tool that helps with:
- **Security Vulnerability Scanning**: Identifies known vulnerabilities in dependencies
- **License Compliance**: Ensures dependencies comply with your license requirements
- **Open Source Compliance**: Manages open source licensing and attribution
- **Policy Enforcement**: Enforces security and licensing policies

## Quick Setup

### 1. Local Development Setup

```bash
# Run the setup script
./scripts/setup_fossa.sh

# Set your FOSSA API key
export FOSSA_API_KEY=your_api_key_here

# Initialize and analyze
fossa init
fossa analyze
fossa test
```

### 2. CI/CD Setup

To enable FOSSA in your CI/CD pipeline:

1. **Get a FOSSA API Key**:
   - Go to [FOSSA API Tokens](https://app.fossa.com/account/settings/integrations/api_tokens)
   - Create a new API token
   - Copy the token

2. **Add to GitHub Secrets**:
   - Go to your repository settings
   - Navigate to "Secrets and variables" → "Actions"
   - Add a new secret named `FOSSA_API_KEY`
   - Paste your API token as the value

3. **Verify Setup**:
   - Push a commit to trigger the CI pipeline
   - Check that FOSSA analysis runs successfully
   - View the FOSSA badge in your README

## Configuration

### FOSSA Configuration File (`.fossa.yml`)

The project includes a comprehensive FOSSA configuration:

```yaml
# Project metadata
name: fl-go
displayName: FL-GO - Federated Learning in Go
projectURL: https://github.com/fedai-oss/fl-go

# Analysis targets
analyze:
  modules:
    - name: fl-go-go-modules
      type: go
      path: .
      options:
        allow-unresolved: false
        allow-nested: true

# Policy configuration
test:
  failOnSeverity: high
  allowlist:
    licenses:
      - MIT
      - Apache-2.0
      - BSD-2-Clause
      - BSD-3-Clause
      - ISC
      - MPL-2.0
      - CC0-1.0
      - Unlicense
      - WTFPL
```

### License Policy

**Allowed Licenses** (permitted for use):
- MIT
- Apache-2.0
- BSD-2-Clause
- BSD-3-Clause
- ISC
- MPL-2.0
- CC0-1.0
- Unlicense
- WTFPL

**Blocked Licenses** (not permitted):
- GPL-1.0, GPL-2.0, GPL-3.0
- AGPL-1.0, AGPL-3.0
- LGPL-2.0, LGPL-2.1, LGPL-3.0

### Security Policy

- **Fail on High Severity**: CI will fail if high-severity vulnerabilities are found
- **Warn on Medium/Low**: Medium and low-severity issues will be reported but won't fail CI
- **Allowlist**: Known false positives can be allowlisted

## Workflow Integration

### CI/CD Pipeline

FOSSA is integrated into the main CI pipeline as a job:

```yaml
# Job 8: FOSSA Dependency Analysis
fossa:
  name: FOSSA Dependency Analysis
  runs-on: ubuntu-latest
  needs: test
  continue-on-error: true
```

### Standalone Workflow

There's also a dedicated FOSSA workflow (`.github/workflows/fossa.yml`) that:
- Runs on push/PR to main/develop branches
- Runs weekly on Mondays at 3 AM UTC
- Generates detailed reports
- Comments on PRs with analysis results

## Reports and Artifacts

### Generated Reports

The CI pipeline generates several reports:

1. **Dependency Report** (`fossa-dependencies.json`):
   - List of all dependencies
   - Vulnerability information
   - License information

2. **License Report** (`fossa-licenses.json`):
   - License compliance status
   - License types and counts
   - Policy violations

3. **Attribution Report** (`fossa-attribution.json`):
   - Required attributions for dependencies
   - Copyright notices

### Accessing Reports

Reports are uploaded as CI artifacts and can be downloaded from:
- GitHub Actions → Workflow Run → Artifacts → `fossa-reports`

## Dashboard Integration

### FOSSA Web Dashboard

View detailed analysis at the FOSSA dashboard:
- **URL**: https://app.fossa.com/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go
- **Features**:
  - Interactive dependency tree
  - Vulnerability details
  - License compliance status
  - Policy management
  - Historical analysis

### Badge Integration

The README includes a FOSSA badge that shows the current status:
```markdown
[![FOSSA](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go)
```

## Troubleshooting

### Common Issues

1. **FOSSA API Key Not Set**:
   ```
   Error: A FOSSA API key was specified, but it is an empty string
   ```
   **Solution**: Add your FOSSA API key to GitHub repository secrets

2. **FOSSA CLI Installation Fails**:
   ```
   platform darwin/arm64 is not supported
   ```
   **Solution**: This is expected on some platforms. FOSSA will work in CI/CD environments.

3. **Policy Violations**:
   ```
   Policy violation: License not allowed
   ```
   **Solution**: Review the dependency and either:
   - Update to a version with an allowed license
   - Add the license to the allowlist (if appropriate)
   - Replace the dependency

### Getting Help

- **FOSSA Documentation**: https://docs.fossa.com/
- **FOSSA Support**: https://fossa.com/support/
- **GitHub Issues**: Create an issue in this repository for FL-GO specific problems

## Best Practices

1. **Regular Updates**: Keep dependencies updated to minimize vulnerabilities
2. **Policy Review**: Regularly review and update license policies
3. **False Positive Management**: Use allowlists for known false positives
4. **Documentation**: Document any policy exceptions or special cases
5. **Team Training**: Ensure team members understand FOSSA policies

## Integration with Other Tools

FOSSA works well with other security tools in the FL-GO project:

- **CodeQL**: Static analysis for security vulnerabilities
- **Gosec**: Go-specific security scanning
- **Govulncheck**: Go vulnerability database checking

Together, these tools provide comprehensive security coverage for the FL-GO project.
