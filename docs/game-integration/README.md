# Blog to Game Integration System

This system automatically converts blog posts into interactive characters for our pixel art game, complete with personality traits, dialogue options, and voice synthesis in both English and Japanese.

## Table of Contents

- [How It Works](#how-it-works)
- [Setup Requirements](#setup-requirements)
- [Documentation](#documentation)
- [Testing](#testing)
- [Troubleshooting](#troubleshooting)

## How It Works

When a new blog post is created in this repository:

1. The Game Content Notifier extracts the raw post data
2. A webhook notification is sent to the game repository
3. The game repository processes the data to generate:
   - Character attributes based on the post content
   - Dialogue options in both English and Japanese
   - Audio clips using ElevenLabs voice synthesis
4. The game is automatically built and deployed with the new character

See the [system architecture diagram](./diagrams/architecture.md) for a visual overview.

## Setup Requirements

### In This Repository (www-jp)

1. **Game Content Notifier Workflow** (`.github/workflows/game-content-notifier.yaml`)
   - Already implemented
   - Extracts post data and sends notification to game repo

2. **Integration with Post Workflow** (`.github/workflows/poster.yaml`)
   - Already implemented
   - Now calls the Game Content Notifier after post creation

3. **Required Secrets**
   - `GAME_REPO_TOKEN`: GitHub token with access to game repo
   - `GAME_REPO_OWNER`: Owner (username or organization) of game repo
   - `GAME_REPO_NAME`: Name of the game repository

### In the Game Repository

See the [Game Implementation Guide](./implementation/game-repo-guide.md) for complete setup instructions.

## Documentation

- [Technical Implementation Details](./implementation/technical-details.md) - Complete system details
- [Secrets Setup Guide](./implementation/secrets-setup.md) - How to configure required secrets
- [System Architecture Diagram](./diagrams/architecture.md) - Visual overview of the system

## Testing

Use the test workflow to verify your setup:

1. Go to Actions → "Test Game Integration"
2. Run the workflow with a valid issue number
3. Check that post content is correctly extracted
4. Verify the webhook would be sent correctly

## Troubleshooting

### Common Issues in www-jp Repository

- **Workflow Not Triggered**: Check that the post has the "post" label and "/post" comment
- **Secret Access Issues**: Verify secrets are correctly set in repository settings
- **Webhook Format Errors**: Check JSON escaping in the webhook payload

### Game Repository Issues

For issues in the game repository, refer to the [Game Implementation Guide](./implementation/game-repo-guide.md).