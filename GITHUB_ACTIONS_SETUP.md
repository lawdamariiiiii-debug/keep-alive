# GitHub Actions Setup Guide

## Overview
The GitHub Actions workflow allows you to run the file keepalive service automatically in the cloud, without needing your own server.

## Setup Steps

### 1. Add Repository Secrets

Go to your GitHub repository settings and add these secrets:

1. Navigate to: **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**
3. Add the following secrets:

| Secret Name | Description | Example |
|------------|-------------|---------|
| `SUPABASE_URL` | Your Supabase project URL | `https://xxxxx.supabase.co` |
| `SUPABASE_KEY` | Your Supabase API key | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` |

### 2. Enable GitHub Actions

1. Go to the **Actions** tab in your repository
2. If prompted, click **"I understand my workflows, go ahead and enable them"**
3. You should now see the **"File Keepalive"** workflow

### 3. Run the Workflow

#### Option A: Manual Trigger (Recommended for First Run)

1. Go to **Actions** tab
2. Click on **"File Keepalive"** workflow
3. Click **"Run workflow"** button
4. Configure options:
   - **Dry run mode**: Check this for a test run (no actual downloads)
   - **Max age days**: Only check files from last N days (default: 30)
   - **Reset state**: Check to start fresh (ignore previous progress)
5. Click **"Run workflow"**

#### Option B: Automatic Schedule

The workflow runs automatically every 5 hours. No action needed!

## Workflow Features

### ✅ Automatic Scheduling
- Runs every 5 hours continuously
- Stays within GitHub Actions free tier limits (2,000 minutes/month)
- Automatic restarts ensure continuous operation

### ✅ State Persistence
- Progress is saved using GitHub Actions cache
- Resumes from where it left off after each run
- No duplicate processing of files

### ✅ Manual Control
- **Dry Run**: Test without downloading files
- **Max Age Days**: Limit to recent files only
- **Reset State**: Start fresh if needed

### ✅ Monitoring & Debugging
- View logs in real-time during execution
- State files saved as artifacts (7-day retention)
- Error logs uploaded on failure

## Usage Examples

### First Time Setup (Test Run)
```
1. Go to Actions → File Keepalive → Run workflow
2. Enable "Dry run mode" ✓
3. Set "Max age days" to 7
4. Click "Run workflow"
5. Check logs to verify configuration
```

### Production Run
```
1. Go to Actions → File Keepalive → Run workflow
2. Leave "Dry run mode" unchecked
3. Set "Max age days" to 30 (or your preference)
4. Click "Run workflow"
5. Let it run automatically every 5 hours
```

### Reset and Start Fresh
```
1. Go to Actions → File Keepalive → Run workflow
2. Enable "Reset state" ✓
3. Click "Run workflow"
4. All previous progress will be cleared
```

## Monitoring

### View Running Workflow
1. Go to **Actions** tab
2. Click on the running workflow
3. Click on the **"keepalive"** job
4. Expand steps to see real-time logs

### Check State Files
1. Go to completed workflow run
2. Scroll to **Artifacts** section
3. Download `keepalive-state-XXXXX` to see progress

### View Logs
- Real-time logs available during execution
- Failed runs upload logs as artifacts
- Logs retained for 7 days

## Troubleshooting

### Workflow Not Showing
- ✅ **Fixed!** Workflow file is now in `.github/workflows/`
- Refresh the Actions tab
- Check that secrets are configured

### "Supabase connection test failed"
- Verify `SUPABASE_URL` secret is correct
- Verify `SUPABASE_KEY` secret is correct
- Check that secrets don't have extra spaces

### "Rate limit reached"
- The workflow automatically handles rate limits
- Increase delay between files if needed
- Consider running less frequently

### Workflow Times Out
- GitHub Actions has 6-hour limit per job
- Workflow is configured for 5h50m timeout
- State is saved, next run will resume

## Cost & Limits

### GitHub Free Tier
- **2,000 minutes/month** for private repos
- **Unlimited** for public repos
- Each run: ~5 hours = 300 minutes
- **~6 runs/month** on free tier (private repos)

### Optimization Tips
1. Use `max-age-days` to limit files processed
2. Run less frequently if needed (edit cron schedule)
3. Use dry-run mode for testing
4. Make repository public for unlimited minutes

## Advanced Configuration

### Change Schedule
Edit `.github/workflows/file-keepalive.yml`:

```yaml
on:
  schedule:
    # Every 5 hours
    - cron: '0 */5 * * *'
    
    # Every 12 hours (more conservative)
    # - cron: '0 */12 * * *'
    
    # Once per day at midnight UTC
    # - cron: '0 0 * * *'
```

### Adjust Timeout
Edit the `timeout-minutes` value:

```yaml
jobs:
  keepalive:
    timeout-minutes: 350  # 5h50m (default)
    # timeout-minutes: 60  # 1 hour (for testing)
```

### Add Delay Between Files
The workflow uses `--once` flag. To add delays, modify the run command:

```yaml
go run . --once --delay-seconds=300
```

## Security Notes

- ✅ Secrets are encrypted by GitHub
- ✅ Secrets are not exposed in logs
- ✅ Only repository admins can view/edit secrets
- ✅ Workflow runs in isolated environment
- ⚠️ Never commit secrets to code

## Support

### Check Workflow Status
```
Repository → Actions → File Keepalive
```

### Download State File
```
Workflow Run → Artifacts → keepalive-state-XXXXX
```

### View Error Logs
```
Failed Run → Artifacts → keepalive-logs-XXXXX
```

## Next Steps

1. ✅ Add secrets (SUPABASE_URL, SUPABASE_KEY)
2. ✅ Run workflow with dry-run mode
3. ✅ Verify logs show correct configuration
4. ✅ Run production workflow
5. ✅ Monitor first few runs
6. ✅ Let it run automatically!

---

**Need Help?** Open an issue on GitHub with:
- Workflow run URL
- Error message from logs
- Configuration used (dry-run, max-age-days, etc.)
