# Secrets Setup Guide

This guide explains how to configure the required secrets for the blog-to-game integration system.

## Required Secrets for www-jp Repository

To enable the integration, add these secrets to the www-jp repository settings:

1. **GAME_REPO_TOKEN**
   - A GitHub Personal Access Token with permission to trigger workflows
   - Required scope: `repo` (specifically `workflow` and `contents`)
   - Used to send webhooks to the game repository

2. **GAME_REPO_OWNER**
   - GitHub username or organization that owns the game repository
   - For example: `devhou-se`

3. **GAME_REPO_NAME**
   - Name of the game repository
   - For example: `pixel-game`

## Creating a GitHub Personal Access Token

1. Go to your GitHub account settings
2. Click on "Developer settings" in the left sidebar
3. Click on "Personal access tokens" then "Tokens (classic)"
4. Click "Generate new token"
5. Give it a descriptive name like "www-jp to game repo integration"
6. Select the `repo` scope (or just the `workflow` and `contents` subscopes)
7. Click "Generate token"
8. Copy the token immediately - you won't be able to see it again!

## Adding Secrets to GitHub Repository

1. Go to your www-jp repository on GitHub
2. Click on "Settings" tab
3. In the left sidebar, click on "Secrets and variables" then "Actions"
4. Click "New repository secret"
5. Add each of the required secrets:
   - Name: `GAME_REPO_TOKEN` | Value: [your token]
   - Name: `GAME_REPO_OWNER` | Value: [owner name]
   - Name: `GAME_REPO_NAME` | Value: [repo name]

## Verifying Secrets

You can verify that your secrets are correctly configured by:

1. Going to the Actions tab in your repository
2. Selecting the "Test Game Integration" workflow
3. Clicking "Run workflow"
4. Entering a valid issue number
5. Checking the workflow output to see if it reports the secrets as available

## Security Considerations

- Keep your GitHub tokens secure - they provide access to your repositories
- Use tokens with the minimum required permissions
- Rotate tokens periodically for better security
- Consider using GitHub Apps instead of PATs for more granular permissions