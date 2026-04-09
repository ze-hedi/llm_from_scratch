# Setup Guide

## Quick Start

### 1. Clone and Setup

```bash
git clone <your-repo-url>
cd opencode-wrapper
chmod +x *.sh test/*.sh
```

### 2. Setup API Keys

```bash
# Copy environment template
cp .env.example .env

# Edit and add your keys
nano .env  # or vim, code, etc.
```

Add your API keys:
```bash
ANTHROPIC_API_KEY=sk-ant-your-key-here
OPENAI_API_KEY=sk-your-key-here
OPENCLAW_API_KEY=your-key-here
```

### 3. Configure OpenClaw

Choose an agent and set it up:

```bash
# Option 1: Use coder agent
cp agent_coder.json ~/.openclaw/openclaw.json

# Option 2: Use researcher agent
cp agent_researcher.json ~/.openclaw/openclaw.json

# Option 3: Use writer agent
cp agent_writer.json ~/.openclaw/openclaw.json
```

### 4. Test Locally

```bash
# Run tests
./test/run_tests.sh

# Should see: "All tests passed!"
```

### 5. Test OpenClaw Wrapper

```bash
# Basic test (requires OpenClaw installed)
./openclaw-wrapper.sh agent --message "Hello, write a Python hello world"

# Check the session file
ls -la .sessions/
cat .sessions/openclaw_session_*.md
```

## GitHub Actions CI/CD Setup

### 1. Push to GitHub

```bash
git init
git add .
git commit -m "Initial commit"
git branch -M main
git remote add origin <your-repo-url>
git push -u origin main
```

### 2. Add GitHub Secrets

1. Go to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Add these secrets one by one:

| Secret Name | Value |
|-------------|-------|
| `ANTHROPIC_API_KEY` | Your Anthropic API key (sk-ant-...) |
| `OPENAI_API_KEY` | Your OpenAI API key (sk-...) |
| `OPENCLAW_API_KEY` | Your OpenClaw API key |

### 3. Verify CI/CD

1. Make a small change to any `.sh` or `.json` file
2. Commit and push:
   ```bash
   git add .
   git commit -m "Test CI/CD"
   git push
   ```
3. Go to **Actions** tab on GitHub
4. Watch the workflow run
5. Should see green checkmark when tests pass

## CI/CD Behavior

The workflow **only runs** when you change:
- Shell scripts (`*.sh`)
- JSON configs (`*.json`)
- Test files (`test/**`)
- Workflow itself (`.github/workflows/**`)

Changes to README, docs, or other files **won't trigger** tests.

## Troubleshooting

### Tests fail locally

```bash
# Check script syntax
bash -n opencode-wrapper.sh
bash -n openclaw-wrapper.sh

# Validate JSON
python3 -m json.tool agent_coder.json
```

### CI/CD fails

1. Check GitHub Actions logs in the **Actions** tab
2. Verify secrets are set correctly (Settings → Secrets)
3. Ensure all JSON files are valid
4. Check that shell scripts have no syntax errors

### Session files not created

```bash
# Check permissions
ls -la *.sh

# Should show: -rwxr-xr-x
# If not, run: chmod +x *.sh
```

### OpenClaw not found

```bash
# Install OpenClaw
npm install -g openclaw@latest

# Verify installation
which openclaw
openclaw --version
```

## Custom Agent Configuration

Create `agent_custom.json`:

```json
{
  "agent": {
    "model": "anthropic/claude-sonnet-4",
    "systemPrompt": "Your custom prompt here",
    "tools": ["bash", "read", "write", "edit"],
    "temperature": 0.7,
    "maxTokens": 4096
  },
  "metadata": {
    "name": "CustomAgent",
    "description": "Your agent description",
    "version": "1.0.0",
    "author": "Your Name"
  }
}
```

Then use it:
```bash
cp agent_custom.json ~/.openclaw/openclaw.json
./openclaw-wrapper.sh agent --message "Your query"
```

## Best Practices

1. **Never commit `.env`** - It's in `.gitignore` by default
2. **Never commit `.sessions/`** - Session logs are local
3. **Validate JSON** before committing - Run `python3 -m json.tool agent_*.json`
4. **Test locally first** - Run `./test/run_tests.sh` before pushing
5. **Use semantic commits** - e.g., "feat: add new agent", "fix: update config"

## Next Steps

- Customize agent configurations for your needs
- Create more specialized agents
- Integrate with your workflow
- Share your agent configs with the team
