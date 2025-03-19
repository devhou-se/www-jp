# Game Repository Implementation Guide

This guide explains how to implement the blog-to-game integration in your Godot game repository.

## Overview

When a new blog post is created in the www-jp repository, our system:
1. Sends a webhook with post content to your game repository
2. Your game repository should then:
   - Generate character data and dialogue based on post content
   - Create audio clips using ElevenLabs
   - Update game files with new character data
   - Build and deploy the game

## Required Components

### 1. GitHub Workflow Files

#### Create `.github/workflows/process-blog-post.yaml`:

```yaml
name: Process Blog Post

on:
  repository_dispatch:
    types: [new_post]
    
jobs:
  process-game-data:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.10"
          
      - name: Install dependencies
        run: pip install -r scripts/requirements.txt
        
      - name: Extract post data
        id: extract-data
        run: |
          echo '${{ toJson(github.event.client_payload) }}' > post_data.json
          POST_ID=$(jq -r '.post_id' post_data.json)
          POST_AUTHOR=$(jq -r '.author' post_data.json)
          echo "post_id=$POST_ID" >> $GITHUB_OUTPUT
          echo "post_author=$POST_AUTHOR" >> $GITHUB_OUTPUT
          
      - name: Generate game data
        run: python scripts/game-data-generator.py
        env:
          POST_DATA_FILE: post_data.json
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          
      - name: Generate audio
        run: python scripts/voice-synthesis.py
        env:
          ELEVENLABS_API_KEY: ${{ secrets.ELEVENLABS_API_KEY }}
          POST_ID: ${{ steps.extract-data.outputs.post_id }}
          POST_AUTHOR: ${{ steps.extract-data.outputs.post_author }}
          
      - name: Upload audio to GCS
        run: python scripts/gcs-uploader.py
        env:
          GCS_BUCKET: ${{ secrets.GCS_BUCKET }}
          GCS_CREDENTIALS: ${{ secrets.GCS_CREDENTIALS }}
          POST_ID: ${{ steps.extract-data.outputs.post_id }}
          
      - name: Update game data
        run: |
          mkdir -p data/characters/
          mkdir -p data/audio/
          cp generated/character-${{ steps.extract-data.outputs.post_id }}.json data/characters/
          cp generated/audio-refs-${{ steps.extract-data.outputs.post_id }}.json data/audio/
          
      - name: Commit changes
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add data/characters/ data/audio/
          git commit -m "Add game data for post #${{ steps.extract-data.outputs.post_id }}"
          git push
```

#### Create `.github/workflows/build-deploy.yaml`:

```yaml
name: Build and Deploy Game

on:
  push:
    branches:
      - main
    paths:
      - 'data/characters/**'
      - 'data/audio/**'
      
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        
      - name: Setup Godot
        uses: chickensoft-games/setup-godot@v1
        with:
          version: 4.2.1
          export-templates: true
          
      - name: Build WASM
        run: |
          mkdir -p build/web
          godot --headless --export-release "Web" build/web/index.html
          
      - name: Deploy to hosting
        uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          command: pages deploy build/web --project-name=game
```

### 2. Python Scripts

#### Create `scripts/requirements.txt`:

```
openai>=1.0.0
google-cloud-storage>=2.7.0
requests>=2.28.0
```

#### Create `scripts/game-data-generator.py`:

This script uses OpenAI to generate character data from the blog post content.

[See full script in the complete implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md#create-scriptsgame-data-generatorpy)

#### Create `scripts/voice-synthesis.py`:

This script uses ElevenLabs to generate audio files for character dialogue.

[See full script in the complete implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md#create-scriptsvoice-synthesispy)

#### Create `scripts/gcs-uploader.py`:

This script uploads audio files to Google Cloud Storage and updates references.

[See full script in the complete implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md#create-scriptsgcs-uploaderpy)

### 3. Godot Scripts

#### Create `scripts/character_manager.gd`:

This script manages character data and audio in the game.

[See full script in the complete implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md#create-scriptscharacter_managergd)

#### Create `scripts/npc_controller.gd`:

This script controls NPC behavior based on character data.

[See full script in the complete implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md#create-scriptsnpc_controllergd)

### 4. Required Directories

Create these directories in your repository:

1. `.github/workflows/`
2. `scripts/`
3. `data/characters/`
4. `data/audio/`

## Required GitHub Secrets

Add these secrets to your game repository:

1. `OPENAI_API_KEY` - API key for OpenAI
2. `ELEVENLABS_API_KEY` - API key for ElevenLabs voice synthesis
3. `GCS_BUCKET` - Name of your Google Cloud Storage bucket
4. `GCS_CREDENTIALS` - JSON credentials for GCS service account (formatted as a JSON string)
5. `CF_API_TOKEN` - Cloudflare API token (if deploying to Cloudflare Pages)
6. `CF_ACCOUNT_ID` - Cloudflare account ID

## Godot Project Setup

1. Set up the CharacterManager as an autoloaded singleton in your Godot project
2. Create NPC scenes that use the NPC Controller script
3. Implement UI elements for dialogue interaction

## Testing

To test if your repository dispatch is working correctly, watch the Actions tab after a blog post is created. You should see the Process Blog Post workflow trigger automatically.

## Additional Information

- Data files will be stored in `data/characters/` and `data/audio/`
- Audio files will be hosted in Google Cloud Storage for lazy loading
- The system supports both English and Japanese dialogue

For complete implementation details, see our [full implementation guide](https://github.com/devhou-se/www-jp/tree/main/game-docs/claude-code-instructions.md).