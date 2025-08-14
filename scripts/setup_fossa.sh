#!/bin/bash

# FOSSA Setup Script for FL-GO
# This script helps set up FOSSA for dependency analysis

set -e

echo "🔧 Setting up FOSSA for FL-GO project..."

# Check if FOSSA CLI is installed
if ! command -v fossa &> /dev/null; then
    echo "📦 Installing FOSSA CLI..."
    # Try the latest version first
    if curl -H 'Cache-Control: no-cache' https://raw.githubusercontent.com/fossas/fossa-cli/master/install-latest.sh | bash; then
        echo "✅ FOSSA CLI v3 installed successfully"
    else
        echo "⚠️  FOSSA CLI installation failed. This is expected on some platforms."
        echo "   FOSSA will still work in CI/CD environments."
        echo "   For local development, you can use the web interface:"
        echo "   https://app.fossa.com/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go"
    fi
else
    echo "✅ FOSSA CLI already installed"
fi

# Check if FOSSA API key is set
if [ -z "$FOSSA_API_KEY" ]; then
    echo "⚠️  FOSSA_API_KEY environment variable not set"
    echo "   Please set your FOSSA API key:"
    echo "   export FOSSA_API_KEY=your_api_key_here"
    echo "   Or add it to your GitHub repository secrets for CI/CD"
    echo ""
    echo "   You can get your API key from: https://app.fossa.com/account/settings/integrations/api_tokens"
else
    echo "✅ FOSSA API key is configured"
fi

# Initialize FOSSA project if not already done
if [ ! -f ".fossa.yml" ]; then
    echo "📝 Creating FOSSA configuration..."
    cat > .fossa.yml << 'EOF'
# FOSSA configuration for FL-GO project
version: 3

name: fl-go
displayName: FL-GO - Federated Learning in Go
projectURL: https://github.com/fedai-oss/fl-go

analyze:
  modules:
    - name: fl-go-go-modules
      type: go
      path: .
      options:
        allow-unresolved: false
        allow-nested: true

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

reports:
  dependencies:
    format: json
    output: fossa-dependencies.json
  licenses:
    format: json
    output: fossa-licenses.json

policy:
  security:
    failOnSeverity: high
  license:
    requireCompliance: true
    allowlist:
      - MIT
      - Apache-2.0
      - BSD-2-Clause
      - BSD-3-Clause
      - ISC
      - MPL-2.0
      - CC0-1.0
      - Unlicense
      - WTFPL
    blocklist:
      - GPL-1.0
      - GPL-2.0
      - GPL-3.0
      - AGPL-1.0
      - AGPL-3.0
      - LGPL-2.0
      - LGPL-2.1
      - LGPL-3.0

ci:
  failOnPolicyViolation: true
  generateReports: true
  commentOnPR: true
EOF
    echo "✅ FOSSA configuration created"
else
    echo "✅ FOSSA configuration already exists"
fi

# Test FOSSA setup
echo "🧪 Testing FOSSA setup..."
if fossa --version &> /dev/null; then
    echo "✅ FOSSA CLI is working"
else
    echo "❌ FOSSA CLI test failed"
    exit 1
fi

echo ""
echo "🎉 FOSSA setup complete!"
echo ""
echo "Next steps:"
echo "1. Set your FOSSA API key: export FOSSA_API_KEY=your_key"
echo "2. Run: fossa init (if not already done)"
echo "3. Run: fossa analyze"
echo "4. Run: fossa test"
echo ""
echo "For CI/CD:"
echo "1. Add FOSSA_API_KEY to your GitHub repository secrets"
echo "2. The CI pipeline will automatically run FOSSA analysis"
echo ""
echo "FOSSA Dashboard: https://app.fossa.com/projects/git%2Bgithub.com%2Ffedai-oss%2Ffl-go"
