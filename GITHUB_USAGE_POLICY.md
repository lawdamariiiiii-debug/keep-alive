# GitHub Actions Usage Policy & Risk Assessment

## ⚠️ Important Considerations

### GitHub's Acceptable Use Policy

GitHub Actions is designed for **CI/CD (Continuous Integration/Continuous Deployment)** - building, testing, and deploying code. Using it for other purposes may violate their Terms of Service.

## Risk Assessment

### 🔴 HIGH RISK Activities (May Lead to Ban)

1. **Cryptocurrency Mining** - Explicitly prohibited
2. **Excessive Resource Usage** - Running 24/7 compute-intensive tasks
3. **Abuse of Free Tier** - Using Actions as free hosting/compute
4. **Network Abuse** - Excessive downloads/uploads, DDoS-like behavior
5. **Violating Third-Party ToS** - If Gofile/Filester prohibit automation

### 🟡 MEDIUM RISK (Current Setup)

Your current workflow has some concerns:

**Concerns:**
- ✅ Running every 5 hours continuously (not typical CI/CD)
- ✅ Downloading large files (bandwidth usage)
- ✅ Using Actions as a "keepalive service" (not its intended purpose)
- ✅ Potentially high bandwidth consumption

**Mitigating Factors:**
- ✅ Not mining cryptocurrency
- ✅ Not running 24/7 (5-hour intervals with breaks)
- ✅ Has legitimate use case (preventing data loss)
- ✅ Not abusing compute resources (mostly I/O)

### 🟢 LOW RISK Alternatives

1. **Self-hosted runner** - Run on your own hardware
2. **Cloud VM** - AWS/GCP/Azure free tier
3. **Home server** - Raspberry Pi or old computer
4. **Scheduled task on personal computer**

## GitHub's Perspective

### What GitHub Cares About:

1. **Compute Abuse** - Are you mining crypto or running heavy compute?
   - ❌ No, you're just downloading files

2. **Resource Fairness** - Are you hogging resources from other users?
   - ⚠️ Potentially, if downloading many large files

3. **Intended Use** - Is this what Actions was designed for?
   - ❌ No, Actions is for CI/CD, not file management

4. **Terms Violation** - Are you violating explicit terms?
   - ⚠️ Gray area - not explicitly prohibited but not intended use

## Recommendations

### Option 1: Reduce Risk on GitHub Actions (Safer)

Modify the workflow to be more conservative:

```yaml
on:
  schedule:
    # Run once per day instead of every 5 hours
    - cron: '0 2 * * *'  # 2 AM UTC daily
```

**Benefits:**
- Less frequent = less resource usage
- More like a "maintenance task"
- Lower chance of triggering abuse detection

**Limitations:**
- Files might expire if not accessed frequently enough

### Option 2: Use GitHub Actions Sparingly (Recommended)

Only use for critical files:

```bash
# Limit to recent files only
--max-age-days=7

# Add longer delays between files
--delay-seconds=600  # 10 minutes
```

**Benefits:**
- Minimal resource usage
- Stays under the radar
- Still provides value

### Option 3: Self-Hosted Runner (Best for Heavy Use)

Set up your own runner:

```bash
# On your own computer/server
1. Go to Settings → Actions → Runners
2. Click "New self-hosted runner"
3. Follow setup instructions
4. Workflow runs on YOUR hardware
```

**Benefits:**
- ✅ No GitHub resource usage
- ✅ No risk of ban
- ✅ Unlimited usage
- ✅ Full control

**Drawbacks:**
- Requires your own hardware
- Must keep computer running

### Option 4: Cloud VM (Most Reliable)

Use a cloud provider's free tier:

**AWS EC2 Free Tier:**
- 750 hours/month free for 12 months
- t2.micro instance
- Perfect for this use case

**Google Cloud Free Tier:**
- e2-micro instance
- Always free (with limits)
- 1GB egress/month

**Oracle Cloud Free Tier:**
- 2 AMD-based VMs
- Always free
- More generous than others

### Option 5: Home Server (Zero Cost)

Run on Raspberry Pi or old computer:

```bash
# One-time setup
./file-keepalive --interval-days=7

# Or use systemd/cron
crontab -e
0 */12 * * * /path/to/file-keepalive --once
```

**Benefits:**
- ✅ Zero ongoing cost
- ✅ No risk of ban
- ✅ Full control

## What Could Trigger a Ban?

### Likely Triggers:
1. **Excessive bandwidth** - Downloading TBs of data
2. **Continuous 24/7 usage** - Never stopping
3. **Multiple accounts** - Trying to bypass limits
4. **Complaints from third parties** - Gofile/Filester reporting abuse
5. **Pattern detection** - Automated systems flagging unusual behavior

### Warning Signs:
- Workflow suddenly disabled
- Email from GitHub about ToS violation
- Account suspended
- Actions minutes exhausted quickly

## Best Practices to Avoid Issues

### 1. Be Conservative
```yaml
# Run less frequently
- cron: '0 0 * * *'  # Once per day

# Limit scope
--max-age-days=7  # Only recent files

# Add delays
--delay-seconds=600  # 10 minutes between files
```

### 2. Monitor Usage
- Check Actions minutes used: Settings → Billing
- Watch for warnings or emails from GitHub
- Keep bandwidth reasonable (<100GB/month)

### 3. Have a Backup Plan
- Keep local copy of important files
- Have alternative hosting ready
- Don't rely solely on GitHub Actions

### 4. Consider Making Repo Public
- Public repos get unlimited Actions minutes
- More transparent = less suspicious
- Community can benefit from your code

### 5. Add Legitimate CI/CD
```yaml
# Add actual CI/CD tasks
- name: Run tests
  run: go test ./...

- name: Build binary
  run: go build .

# Then do keepalive as "maintenance"
```

## My Recommendation

### For Your Use Case:

**Short Term (Testing):**
- ✅ Use GitHub Actions with conservative settings
- ✅ Run once per day maximum
- ✅ Limit to recent files (7 days)
- ✅ Monitor for any warnings

**Long Term (Production):**
- ✅ Set up self-hosted runner OR
- ✅ Use cloud VM free tier OR
- ✅ Run on home server/Raspberry Pi

**Why?**
- GitHub Actions is not designed for this
- Risk of ban increases over time
- Self-hosted is more reliable and ethical
- Cloud VMs are cheap/free and purpose-built

## Legal Disclaimer

I am not a lawyer. This is not legal advice. Review GitHub's Terms of Service:
- https://docs.github.com/en/site-policy/github-terms/github-terms-of-service
- https://docs.github.com/en/actions/learn-github-actions/usage-limits-billing-and-administration

**Key Quote from GitHub ToS:**
> "You may not use GitHub Actions if such use [...] places excessive burden on GitHub's systems"

## Conclusion

### Risk Level: 🟡 MEDIUM

**Will you get banned immediately?** Probably not.

**Could you get banned eventually?** Possibly, especially if:
- You download large amounts of data
- You run continuously for months
- GitHub detects abuse patterns
- Third parties complain

**Safest Approach:**
1. Test with GitHub Actions (conservative settings)
2. Transition to self-hosted runner or cloud VM
3. Keep GitHub Actions as backup only

### Final Recommendation:

**Use GitHub Actions for testing and emergency use only. For production, use a self-hosted solution.**

This keeps you ethical, avoids ToS violations, and ensures reliability.

---

**Questions?** Consider:
- How much data are you downloading per month?
- How critical is this service to you?
- Do you have a computer that can run 24/7?
- Can you afford $5-10/month for a cloud VM?

Choose the solution that matches your needs and risk tolerance.
