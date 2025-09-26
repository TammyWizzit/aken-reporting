#!/bin/bash

# Setup script for SSH authentication with wizzitdigital repositories
# This script helps developers configure Git to use SSH for private repositories

set -e

echo "🔧 Setting up SSH authentication for wizzitdigital repositories..."

# Check if SSH key exists
if [ ! -f ~/.ssh/id_ed25519 ] && [ ! -f ~/.ssh/id_rsa ]; then
    echo "❌ No SSH key found. Please generate one first:"
    echo "   ssh-keygen -t ed25519 -C \"your_email@example.com\""
    echo "   Then add it to your GitHub account."
    exit 1
fi

# Configure Git to use SSH for wizzitdigital
echo "📝 Configuring Git to use SSH for wizzitdigital repositories..."
git config --global url."git@github.com:wizzitdigital/".insteadOf "https://github.com/wizzitdigital/"

# Test SSH connection
echo "🔍 Testing SSH connection to GitHub..."
if ssh -T git@github.com 2>&1 | grep -q "successfully authenticated"; then
    echo "✅ SSH authentication successful!"
else
    echo "❌ SSH authentication failed. Please check your SSH key setup."
    echo "   Make sure your SSH key is added to your GitHub account."
    exit 1
fi

echo "🎉 Setup complete! You can now access wizzitdigital repositories via SSH."
echo ""
echo "Next steps:"
echo "1. Add the wizzit-logger dependency: go get github.com/wizzitdigital/wizzit-logger@latest"
echo "2. Run: go mod tidy"
echo "3. Verify: go list -m github.com/wizzitdigital/wizzit-logger"


