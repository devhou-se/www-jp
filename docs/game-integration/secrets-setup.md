# Game Integration Secrets Setup Guide

This guide explains how to set up the required secrets for the blog-to-game integration.

## Secrets for www-jp Repository

These secrets need to be added to the www-jp repository to enable the Game Content Notifier workflow:

| Secret Name | Description | Example |
|-------------|-------------|---------|
| `GAME_REPO_TOKEN` | GitHub Personal Access Token with permissions to trigger workflows in the game repo | `ghp_1a2b3c4d5e6f7g8h9i0j` |
| `GAME_REPO_OWNER` | GitHub username or organization that owns the game repo | `yourusername` |
| `GAME_REPO_NAME` | Name of the game repository | `pixel-game` |

## Creating a GitHub Personal Access Token

1. Go to your GitHub account settings
2. Select "Developer settings" from the sidebar
3. Click "Personal access tokens" → "Fine-grained tokens"
4. Click "Generate new token"
5. Give it a descriptive name like "Game Integration Token"
6. Set the expiration period (e.g., 1 year)
7. Select the repository you want to access
8. Under "Repository permissions":
   - "Contents": Read-only (as we only need to trigger a workflow)
   - "Metadata": Read-only
   - "Actions": Read and write (to create repository dispatch events)
9. Click "Generate token"
10. Copy the token immediately (you won't be able to see it again)

## Adding Secrets to GitHub Repository

1. Go to your repository on GitHub
2. Click on "Settings"
3. From the sidebar, click "Secrets and variables" → "Actions"
4. Click "New repository secret"
5. Enter the secret name and value
6. Click "Add secret"
7. Repeat for each required secret

## Testing Your Secrets

Use the built-in test workflow to verify your secrets are configured correctly:

1. Go to Actions → "Test Game Integration"
2. Click "Run workflow"
3. Enter an issue number to test with (e.g., "1")
4. Review the logs to check:
   - The workflow can access the secrets
   - Post content is correctly extracted
   - The webhook would be properly formatted

Example test output:
```
Secret GAME_REPO_TOKEN available: true
Secret GAME_REPO_OWNER available: true
Secret GAME_REPO_NAME available: true
Would send webhook to: https://api.github.com/repos/yourusername/game/dispatches
```

## Security Best Practices

### Token Permissions

When creating your GitHub Personal Access Token, follow the principle of least privilege:

1. **Only grant necessary permissions**: The token only needs "Actions" write access to create repository dispatch events
2. **Limit repository access**: Select only the game repository, not all repositories
3. **Set a reasonable expiration date**: Don't create tokens that never expire

### Repository Protection

In the game repository:

1. **Branch protection**: Enable branch protection rules for the main branch
2. **Required reviews**: Consider requiring pull request reviews for changes
3. **Status checks**: Require status checks to pass before merging

### Secret Handling

1. **No secret logging**: Never log or expose secret values in workflow outputs
2. **Regular rotation**: Periodically rotate the GitHub token and other secrets
3. **Minimum scope**: Use repository-level secrets rather than organization secrets when possible

## Troubleshooting Secret Issues

### Token Not Working

If the webhook isn't being received by the game repository:

1. Check the token has the correct permissions
2. Verify the token hasn't expired
3. Ensure the token is added as a secret correctly (check for whitespace)

### Repository Access Issues

If the webhook request fails with a 404:

1. Verify the `GAME_REPO_OWNER` and `GAME_REPO_NAME` are correct
2. Check that the token has access to the repository
3. Ensure the repository exists and is not private (or that the token has access if it is private)

### Test Your Setup

Run this command locally to test the webhook (replace placeholders with actual values):

```bash
curl -X POST \
  -H "Authorization: token YOUR_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/dispatches \
  -d '{"event_type": "test_event"}'
```

A 204 status code indicates success.