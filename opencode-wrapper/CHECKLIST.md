# Pre-Push Checklist

Use this checklist before pushing to GitHub to ensure everything works.

## ✅ Local Setup

- [ ] All scripts are executable (`chmod +x *.sh test/*.sh`)
- [ ] `.env.example` exists (template for API keys)
- [ ] `.gitignore` includes `.env`, `.sessions/`, `test/results/`
- [ ] No `.env` file in git (`git status` should not show it)

## ✅ Code Quality

- [ ] Shell scripts have valid syntax
  ```bash
  bash -n opencode-wrapper.sh
  bash -n openclaw-wrapper.sh
  bash -n test/run_tests.sh
  ```

- [ ] All JSON files are valid
  ```bash
  python3 -m json.tool agent_coder.json
  python3 -m json.tool agent_researcher.json
  python3 -m json.tool agent_writer.json
  ```

- [ ] All tests pass locally
  ```bash
  ./test/run_tests.sh
  # Should output: "All tests passed!"
  ```

## ✅ Documentation

- [ ] README.md is up to date
- [ ] SETUP_GUIDE.md has clear instructions
- [ ] Agent configs have proper metadata
- [ ] No TODO comments in code
- [ ] Examples work as described

## ✅ CI/CD Preparation

### Before First Push

- [ ] Repository created on GitHub
- [ ] GitHub Secrets configured:
  - [ ] `ANTHROPIC_API_KEY`
  - [ ] `OPENAI_API_KEY`
  - [ ] `OPENCLAW_API_KEY`

### Test CI/CD

- [ ] Push to GitHub
  ```bash
  git add .
  git commit -m "Initial commit"
  git push -u origin main
  ```

- [ ] Check Actions tab on GitHub
- [ ] Verify workflow runs automatically
- [ ] All CI/CD tests pass (green checkmark)

## ✅ Functional Testing

- [ ] OpenClaw wrapper creates session files
  ```bash
  ./openclaw-wrapper.sh agent --message "test"
  ls .sessions/
  ```

- [ ] Config extraction works
  ```bash
  cat .sessions/openclaw_session_*.md
  # Should show model, prompt, tools
  ```

- [ ] Agent configs can be loaded
  ```bash
  cp agent_coder.json ~/.openclaw/openclaw.json
  # Should not error
  ```

## ✅ Security

- [ ] No API keys in code
- [ ] No API keys in git history (`git log --all -p | grep -i "sk-"`)
- [ ] `.env` in `.gitignore`
- [ ] Only `.env.example` committed
- [ ] GitHub Secrets set (not visible in Actions logs)

## ✅ Final Checks

- [ ] Run full test suite one more time
  ```bash
  ./test/run_tests.sh
  ```

- [ ] Check git status is clean
  ```bash
  git status
  # No unwanted files staged
  ```

- [ ] Review commit message
  ```bash
  git log -1 --oneline
  # Should be descriptive
  ```

## ✅ Post-Push Verification

After pushing to GitHub:

- [ ] Go to Actions tab
- [ ] Watch workflow run
- [ ] All jobs succeed (green checkmarks)
- [ ] No secret leaks in logs
- [ ] Test artifacts uploaded successfully

## Quick Commands

```bash
# Full validation in one go
bash -n *.sh test/*.sh && \
python3 -m json.tool agent_*.json && \
./test/run_tests.sh && \
echo "✅ All checks passed!"

# If all pass, you're ready to push!
git add .
git commit -m "Your commit message"
git push
```

## Troubleshooting

### Tests fail
1. Read error messages carefully
2. Fix the issue
3. Run tests again
4. Don't push until tests pass

### CI/CD fails
1. Check Actions logs on GitHub
2. Look for specific error
3. Fix locally
4. Test locally
5. Push fix

### Secrets not working
1. Verify secrets are set in GitHub Settings
2. Check secret names match workflow
3. Re-save secrets if needed
4. Re-run workflow

---

**Remember**: Only push when ALL checkboxes are checked! ✅
