# Game Integration Technical Documentation

This document provides detailed technical information on how the blog-to-game integration system works.

## System Architecture

The integration consists of components in two repositories:

1. **www-jp repository** (this repository)
   - Handles blog post creation
   - Extracts raw post content
   - Sends webhook to game repository

2. **Game repository**
   - Processes post data
   - Generates character data and dialogue
   - Creates audio with ElevenLabs
   - Builds and deploys the game

## Workflow Details

### Game Content Notifier Workflow

**File**: `.github/workflows/game-content-notifier.yaml`  
**Trigger**: Called by the Post workflow when a new post is created  
**Purpose**: Extract post data and send webhook to game repository

This workflow:
1. Extracts raw post content (title, author, body, date)
2. Formats the data as JSON
3. Sends a repository_dispatch webhook to the game repo

```yaml
name: Game Content Notifier

on:
  workflow_call:
    inputs:
      issue_number:
        required: true
        type: string
        
jobs:
  notify-game-repo:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.10"
          
      - name: Get post content
        id: get-content
        run: |
          POST_BODY=$(gh issue view ${{ inputs.issue_number }} --json body -q .body)
          POST_AUTHOR=$(gh issue view ${{ inputs.issue_number }} --json author -q .author.login)
          POST_TITLE=$(gh issue view ${{ inputs.issue_number }} --json title -q .title)
          POST_DATE=$(gh issue view ${{ inputs.issue_number }} --json createdAt -q .createdAt)
          
          # Clean content for JSON
          POST_BODY_ESCAPED=$(echo "$POST_BODY" | jq -aRs .)
          POST_TITLE_ESCAPED=$(echo "$POST_TITLE" | jq -aRs .)
          
          # Create payload file
          cat > payload.json << EOF
          {
            "client_payload": {
              "post_id": "${{ inputs.issue_number }}",
              "title": $POST_TITLE_ESCAPED,
              "author": "$POST_AUTHOR",
              "date": "$POST_DATE",
              "content": $POST_BODY_ESCAPED
            }
          }
          EOF
          
          cat payload.json
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          
      - name: Trigger game repo workflow
        run: |
          curl -X POST \
            -H "Authorization: token ${{ secrets.GAME_REPO_TOKEN }}" \
            -H "Accept: application/vnd.github.v3+json" \
            https://api.github.com/repos/${{ secrets.GAME_REPO_OWNER }}/${{ secrets.GAME_REPO_NAME }}/dispatches \
            --data @payload.json
```

### Integration with Post Workflow

**File**: `.github/workflows/poster.yaml`  
**Changes**: Added a step to call the Game Content Notifier workflow

```yaml
- name: "notify game repo"
  if: steps.check-comment.outputs.valid == 'true'
  uses: ./.github/workflows/game-content-notifier.yaml
  with:
    issue_number: ${{ github.event.issue.number }}
  secrets: inherit
```

### Test Workflow

**File**: `.github/workflows/test-game-integration.yaml`  
**Purpose**: Test the integration without creating a real post

This workflow:
1. Extracts post data from a specified issue
2. Shows what would be sent to the game repo
3. Can optionally send a real webhook

## Implementation Details

### JSON Escaping

Post content is properly escaped to ensure valid JSON:

```bash
POST_BODY_ESCAPED=$(echo "$POST_BODY" | jq -aRs .)
POST_TITLE_ESCAPED=$(echo "$POST_TITLE" | jq -aRs .)
```

### GitHub API Integration

The webhook uses the GitHub API to send a repository dispatch event:

```bash
curl -X POST \
  -H "Authorization: token ${{ secrets.GAME_REPO_TOKEN }}" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/${{ secrets.GAME_REPO_OWNER }}/${{ secrets.GAME_REPO_NAME }}/dispatches \
  --data @payload.json
```

### Webhook Payload Format

The notification to the game repository includes:

```json
{
  "client_payload": {
    "post_id": "123",
    "title": "My Day in Tokyo",
    "author": "bailey",
    "date": "2025-03-19T12:00:00Z",
    "content": "Today I visited several interesting places..."
  }
}
```

## Required Secrets

To enable the integration, add these secrets to the www-jp repository:

- `GAME_REPO_TOKEN`: GitHub Personal Access Token with permission to trigger workflows
- `GAME_REPO_OWNER`: GitHub username or organization that owns the game repo
- `GAME_REPO_NAME`: Name of the game repository

## Testing the Integration

To test the integration:

1. Go to Actions → "Test Game Integration"
2. Click "Run workflow"
3. Enter an issue number to test with
4. Review the logs to ensure post content is correctly extracted

You can also test with a real post:

1. Create a new GitHub issue with the "post" label
2. Add a comment "/post" to trigger the workflow
3. Check the GitHub Actions tab to verify the Game Content Notifier runs
4. Verify in the game repository that the webhook was received

## Error Handling

The workflow will fail if:
- The issue number is invalid
- Required secrets are missing
- The webhook request fails

## Security Considerations

- The `GAME_REPO_TOKEN` should have minimal permissions (only repository_dispatch)
- Post content is sent as-is, so sensitive information should be avoided in blog posts
- Consider implementing content filtering in the game repository

## Implementation Notes

- The blog repo only handles notification - no game-specific logic
- All game data processing happens in the game repository
- Audio files are stored in Google Cloud Storage for lazy loading