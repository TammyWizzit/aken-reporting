# Developer Setup Guide

## SSH Authentication for Private Repositories

This project uses the `wizzit-logger` package from a private GitHub repository. To access this repository, you need to configure SSH authentication.

### Prerequisites

1. **SSH Key Setup**: Ensure you have an SSH key pair generated and added to your GitHub account
2. **GitHub Access**: Make sure you have access to the `wizzitdigital/wizzit-logger` repository

### SSH Configuration

#### 1. Generate SSH Key (if you don't have one)
```bash
ssh-keygen -t ed25519 -C "your_email@example.com"
```

#### 2. Add SSH Key to GitHub
1. Copy your public key:
   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```
2. Go to GitHub → Settings → SSH and GPG keys → New SSH key
3. Paste your public key and save

#### 3. Configure Git to Use SSH for wizzitdigital
Add the following to your `~/.ssh/config` file:

```
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519

Host github.com-wizzitdigital
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519
```

#### 4. Test SSH Connection
```bash
ssh -T git@github.com
```

You should see: `Hi username! You've successfully authenticated, but GitHub does not provide shell access.`

### Project Setup

#### 1. Clone the Repository
```bash
git clone git@github.com:efinance/aken-reporting-service.git
cd aken-reporting-service
```

#### 2. Configure Git for wizzitdigital
```bash
git config --global url."git@github.com:wizzitdigital/".insteadOf "https://github.com/wizzitdigital/"
```

#### 3. Add wizzit-logger Dependency
After setting up SSH authentication, add the dependency to your project:

```bash
# Add the dependency (replace with actual version/tag from the repository)
go get github.com/wizzitdigital/wizzit-logger@latest

# Or if you know the specific version/tag:
go get github.com/wizzitdigital/wizzit-logger@v1.0.0
```

#### 4. Install Dependencies
```bash
go mod tidy
go mod download
```

#### 5. Verify Installation
```bash
go list -m github.com/wizzitdigital/wizzit-logger
```

### Troubleshooting

#### Issue: "Permission denied (publickey)"
- Ensure your SSH key is added to GitHub
- Test SSH connection: `ssh -T git@github.com`
- Check SSH agent: `ssh-add -l`

#### Issue: "repository not found"
- Verify you have access to the `wizzitdigital/wizzit-logger` repository
- Contact your team lead to request access

#### Issue: "go: module github.com/wizzitdigital/wizzit-logger: git ls-remote -q origin in /path/to/module: exit status 128"
- Ensure Git is configured to use SSH for wizzitdigital repositories
- Run: `git config --global url."git@github.com:wizzitdigital/".insteadOf "https://github.com/wizzitdigital/"`

### Security Notes

- **Never commit personal access tokens** to version control
- **Use SSH keys** for authentication instead of tokens
- **Rotate SSH keys** regularly for security
- **Use different SSH keys** for different organizations if needed

### Alternative: Using Personal Access Tokens (Not Recommended)

If SSH setup is not possible, you can use a personal access token, but this is **not recommended** for security reasons:

1. Create a personal access token in GitHub
2. Configure Git:
   ```bash
   git config --global url."https://YOUR_TOKEN@github.com/wizzitdigital/".insteadOf "https://github.com/wizzitdigital/"
   ```

**Warning**: This approach exposes your token and should only be used temporarily.
