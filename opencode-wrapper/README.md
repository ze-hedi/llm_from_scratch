# Session Wrappers for OpenCode & OpenClaw

Lightweight bash wrappers that automatically save session metadata to markdown files for both OpenCode and OpenClaw.

## Overview

These wrapper scripts launch OpenCode or OpenClaw while capturing session information. When the CLI exits, they automatically create a timestamped markdown file in the `.sessions` directory with session metadata.

## Features

- **Automatic Session Logging**: Creates a `.sessions` directory and saves session information on exit
- **Timestamped Records**: Each session is saved with a unique timestamp (`session_YYYYMMDD_HHMMSS.md`)
- **Clean Exit Handling**: Properly handles EXIT, INT, and TERM signals to ensure session data is saved
- **Transparent Operation**: Passes all arguments directly to the CLI
- **Agent Configurations**: Pre-configured agents for different tasks (coder, researcher, writer)
- **CI/CD Integration**: Automated testing with GitHub Actions
- **Config Extraction**: Automatically captures model, system prompt, and tools used

## Installation

1. Clone or download this repository:
```bash
git clone <repository-url>
cd opencode-wrapper
```

2. Make the scripts executable:
```bash
chmod +x opencode-wrapper.sh
chmod +x openclaw-wrapper.sh
```

3. (Optional) Add aliases to your shell configuration:
```bash
# Add to ~/.bashrc or ~/.zshrc
alias opencode='~/path/to/opencode-wrapper/opencode-wrapper.sh'
alias openclaw='~/path/to/opencode-wrapper/openclaw-wrapper.sh'
```

## Usage

### OpenCode Wrapper

Run the wrapper just like you would run OpenCode:

```bash
./opencode-wrapper.sh [arguments]
```

Or if you've set up the alias:

```bash
opencode [arguments]
```

### OpenClaw Wrapper

Run the wrapper just like you would run OpenClaw:

```bash
./openclaw-wrapper.sh [arguments]
```

Or if you've set up the alias:

```bash
openclaw [arguments]
```

All arguments are passed directly to the respective CLI.

## Agent Configurations

The repository includes pre-configured agent profiles for different tasks:

### Available Agents

1. **Coder** (`agent_coder.json`)
   - Expert software engineer
   - Tools: bash, read, write, edit, glob, grep
   - Ideal for: Code development, debugging, refactoring

2. **Researcher** (`agent_researcher.json`)
   - Research assistant
   - Tools: bash, read, write, webfetch, grep, glob
   - Ideal for: Information gathering, analysis, documentation research

3. **Writer** (`agent_writer.json`)
   - Technical writer
   - Tools: read, write, edit, glob
   - Ideal for: Documentation, content creation, technical writing

### Using Agent Configs

Create your OpenClaw config referencing an agent:

```bash
# Copy an agent config to your OpenClaw config
cp agent_coder.json ~/.openclaw/openclaw.json

# Or create custom config based on agent templates
cat agent_researcher.json | jq '.agent' > ~/.openclaw/openclaw.json
```

### Creating Custom Agents

Create your own `agent_<name>.json`:

```json
{
  "agent": {
    "model": "anthropic/claude-sonnet-4",
    "systemPrompt": "Your custom system prompt here",
    "tools": ["bash", "read", "write"],
    "temperature": 0.7,
    "maxTokens": 4096
  },
  "metadata": {
    "name": "CustomAgent",
    "description": "Description of your agent",
    "version": "1.0.0"
  }
}
```

## Session Files

Session metadata is saved to `.sessions/` in your current working directory:
- OpenCode sessions: `.sessions/session_YYYYMMDD_HHMMSS.md`
- OpenClaw sessions: `.sessions/openclaw_session_YYYYMMDD_HHMMSS.md`

Example OpenCode session file:
```markdown
# OpenCode Session - 2026-04-09 14:30:45

**Working Directory:** /home/user/project

**Session Duration:** Started at 20260409_143045

---

## Session Summary

This session was automatically saved by opencode-wrapper.
```

Example OpenClaw session file:
```markdown
# OpenClaw Session - 2026-04-09 14:30:45

**Working Directory:** /home/user/project

**Session Duration:** Started at 20260409_143045

---

## Session Summary

This session was automatically saved by openclaw-wrapper.
```

## Requirements

- Bash shell
- OpenCode CLI installed and available in PATH (for opencode-wrapper.sh)
- OpenClaw CLI installed and available in PATH (for openclaw-wrapper.sh)
- Python 3 (for JSON validation in tests)
- Node.js 20+ (for CI/CD)

## Testing

### Local Testing

Run the test suite locally:

```bash
./test/run_tests.sh
```

This will:
- Validate shell script syntax
- Validate JSON configurations
- Test session directory creation
- Test config extraction

### CI/CD

The repository includes GitHub Actions workflow that:
- Runs automatically on push/PR when code files change
- Only triggers on changes to `.sh`, `.json`, `test/`, or workflow files
- Uses **GitHub Secrets** for API keys (never exposed in logs)
- Validates all scripts and configs
- Runs the full test suite

### Setting Up CI/CD

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Add the following secrets:
   - `ANTHROPIC_API_KEY` - Your Anthropic API key
   - `OPENAI_API_KEY` - Your OpenAI API key (if using)
   - `OPENCLAW_API_KEY` - Your OpenClaw API key (if needed)

4. The workflow will automatically run on the next push

**Security**: Only repository owners can view/edit secrets. They are never exposed in logs or pull requests from forks.

## Environment Variables

Copy `.env.example` to `.env` for local testing:

```bash
cp .env.example .env
# Edit .env and add your actual API keys
```

**Important**: Never commit `.env` to version control. It's already in `.gitignore`.

## Project Structure

```
.
├── .github/
│   └── workflows/
│       └── test.yml           # CI/CD workflow
├── test/
│   ├── run_tests.sh          # Test runner
│   └── results/              # Test results (generated)
├── .sessions/                # Session logs (generated)
├── agent_coder.json          # Coder agent config
├── agent_researcher.json     # Researcher agent config
├── agent_writer.json         # Writer agent config
├── opencode-wrapper.sh       # OpenCode wrapper
├── openclaw-wrapper.sh       # OpenClaw wrapper
├── .env.example              # Environment template
├── .gitignore               # Git ignore rules
└── README.md                # This file
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

When contributing:
1. Ensure all tests pass locally: `./test/run_tests.sh`
2. Validate JSON configs: `python3 -m json.tool agent_*.json`
3. Follow existing code style
4. Update documentation as needed
