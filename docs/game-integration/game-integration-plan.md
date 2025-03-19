# Game Integration Plan: Blog Post to Game Data Synthesis

## Overview
This plan outlines how to automatically convert blog post content into structured game data for a Godot-based pixel art game featuring virtual versions of blog authors. The game will use AI to synthesize post content into character attributes, dialogue options, and voice clips in both English and Japanese, deployed as a WASM build running in browsers.

## Current System Analysis
The current workflow:
1. Blog posts are created via GitHub issues with the "post" label
2. A GitHub Action is triggered by a "/post" comment
3. The content is processed by `poster.py` to create a Markdown file
4. If translation is requested, `translator.py` translates the content to Japanese using OpenAI's API
5. The post is committed to the repository and deployed

## Game Integration Architecture

### New Components to Create

#### 1. Game Data Processor & Generator
- **Function**: Extract raw post data early in workflow and generate game-compatible structured data
- **Implementation**: Python script using OpenAI API with structured prompting
- **Input**: Raw blog post content from GitHub issue
- **Output**: JSON data structure with character attributes and dialogue

```python
# Example output structure
{
  "character": "Bailey",
  "post_id": 123,
  "date": "2025-03-19",
  "personality_traits": {
    "mood": "excited",
    "energy": 0.8,
    "sociability": 0.7
  },
  "movement_patterns": {
    "speed": 0.65,
    "wandering_radius": 0.4
  },
  "dialogue": {
    "highlights": [
      "Today I solved that tricky API integration problem!",
      "The team celebrated with pizza afterward."
    ],
    "drinking_activities": [
      "Had some craft beers with the team at the local pub."
    ],
    "funny_moments": [
      "Accidentally wrote 'console.log' in a Python file, old habits die hard."
    ]
  },
  "japanese_dialogue": {
    "highlights": [
      "今日はあの難しいAPI統合問題を解決しました！",
      "その後、チームはピザで祝いました。"
    ],
    "drinking_activities": [
      "地元のパブでチームとクラフトビールを飲みました。"
    ],
    "funny_moments": [
      "うっかりPythonファイルに'console.log'と書いてしまいました。古い習慣は死ににくいです。"
    ]
  }
}
```

#### 2. Voice Synthesis Service
- **Function**: Convert text dialogue into audio clips using ElevenLabs API
- **Implementation**: Python script to process dialogue snippets
- **Input**: Text dialogue in both English and Japanese
- **Output**: Audio files stored in Google Cloud Storage (GCS)

#### 3. Game Content Repository Update
- **Function**: Push generated data to the game repository
- **Implementation**: GitHub Action that commits JSON data to the game repo
- **Input**: Generated game data and GCS references
- **Output**: Commit to game repository and trigger for game build

#### 4. Game Integration Module (in Godot)
- **Function**: Code inside the Godot game to utilize character data
- **Implementation**: GDScript code for character behavior and dialogue
- **Features**: Character behavior controls, dialogue UI, audio lazy loading from GCS

### GitHub Actions Workflow

#### In www-jp repo:
1. **Post Processing Action**
   - Triggered by existing post creation workflow ("/post" comment)
   - Extracts raw post content early in the process
   - Runs Game Data Generator in parallel with regular post processing

2. **Game Data Synthesis Action**
   - **Trigger**: After raw post content extraction
   - **Steps**:
     - Generate structured character data using OpenAI
     - Create dialogue in both English and Japanese
     - Generate audio files using ElevenLabs API
     - Upload audio files to Google Cloud Storage
     - Push JSON data and GCS references to game repository

#### In game repo:
3. **Game Build Action**
   - **Trigger**: Commit from Game Data Synthesis Action
   - **Steps**:
     - Pull latest game code with new character data
     - Build Godot project to WASM
     - Deploy WASM build to web hosting

## Implementation Roadmap

### Phase 1: Data Extraction & Structure (1-2 weeks)
- Define complete JSON schema for character data
- Create Game Data Processor for raw post extraction
- Implement GitHub Action to trigger at appropriate point in workflow
- Test with existing blog posts

### Phase 2: AI Data Generation (1-2 weeks)
- Develop prompt engineering for OpenAI to convert post text to structured data
- Implement dialogue generation in both languages
- Create test pipeline to validate data quality
- Integrate with GitHub Actions workflow

### Phase 3: Voice Synthesis (1 week)
- Implement Voice Synthesis service with ElevenLabs
- Set up Google Cloud Storage bucket for audio assets
- Create naming convention and reference system for audio files
- Test audio generation in both languages

### Phase 4: Game Repository Integration (1 week)
- Develop GitHub Action for pushing content to game repo
- Implement webhook or workflow_dispatch for triggering game builds
- Test end-to-end content flow from blog post to game repo

### Phase 5: Godot Integration (2-3 weeks)
- Develop character controller using generated movement data
- Create dialogue UI system with language switching
- Implement audio asset lazy loading from GCS
- Test character behavior and dialogue systems

### Phase 6: Testing & Optimization (1-2 weeks)
- End-to-end testing of full pipeline
- Optimize asset loading and caching
- Refine AI prompts for better character data

## Technical Implementation Details

### GitHub Actions Configuration

```yaml
# In www-jp repo
name: Game Data Generation

on:
  workflow_run:
    workflows: ["Post"]
    types:
      - in_progress

jobs:
  generate-game-data:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.10'
      - name: Install dependencies
        run: pip install -r python/game-requirements.txt
      - name: Generate game data
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          ELEVENLABS_API_KEY: ${{ secrets.ELEVENLABS_API_KEY }}
          GCS_BUCKET: ${{ secrets.GCS_BUCKET }}
          ISSUE_NUMBER: ${{ github.event.workflow_run.pull_requests[0].number }}
        run: python python/game-data-generator.py $ISSUE_NUMBER
      - name: Push to game repo
        env:
          GAME_REPO_TOKEN: ${{ secrets.GAME_REPO_TOKEN }}
        run: python python/game-repo-updater.py
```

### Google Cloud Storage Setup
- Create dedicated bucket for game audio assets
- Set appropriate CORS headers for web game access
- Implement lifecycle policies for managing old assets
- Set up authentication for GitHub Actions

### Game (Godot) Implementation

```gdscript
# Character controller example in GDScript
extends CharacterBody2D

var character_data: Dictionary
var audio_clips: Dictionary = {}
var current_dialogue: String = ""
var language: String = "english" # or "japanese"

func _ready():
    # Load character data from JSON
    var file = File.new()
    file.open("res://data/characters/bailey.json", File.READ)
    character_data = JSON.parse(file.get_as_text()).result
    file.close()
    
    # Set movement patterns
    $MovementController.speed = character_data.movement_patterns.speed * 100
    $MovementController.wander_radius = character_data.movement_patterns.wandering_radius * 200
    
    # Set up dialogue options
    update_dialogue_options()

func update_dialogue_options():
    var dialogue_menu = $DialogueMenu
    dialogue_menu.clear()
    
    # Add dialogue categories as menu items
    for category in character_data.dialogue.keys():
        dialogue_menu.add_item(category)

func on_dialogue_selected(category: String):
    var dialogues = character_data.dialogue[category] if language == "english" else character_data.japanese_dialogue[category]
    if dialogues.size() > 0:
        # Select random dialogue from category
        current_dialogue = dialogues[randi() % dialogues.size()]
        $DialogueLabel.text = current_dialogue
        
        # Lazy load audio if not cached
        var audio_key = "%s_%s_%s" % [character_data.character, language, category]
        if not audio_clips.has(audio_key):
            var audio_url = "https://storage.googleapis.com/game-assets/%s.mp3" % audio_key
            $AudioLoader.load_audio(audio_url, audio_key)
        else:
            $AudioPlayer.stream = audio_clips[audio_key]
            $AudioPlayer.play()

func on_audio_loaded(audio_stream, key):
    audio_clips[key] = audio_stream
    $AudioPlayer.stream = audio_stream
    $AudioPlayer.play()
```

## Additional Considerations

### Data Privacy & Filtering
- Implement content filtering to ensure game-appropriate dialogue
- Create an approval system for sensitive content
- Provide opt-out mechanism for authors

### Performance Optimization
- Implement asset bundling for character data
- Use compression for audio files to reduce storage/bandwidth
- Implement caching for frequently accessed character data

### Multilingual Support
- Ensure proper handling of Japanese text in Godot UI
- Create language switching UI in game
- Consider adding language preference setting

### Cross-Character Interaction
- Generate potential character interactions based on mentions in posts
- Create dialogue options for characters to discuss shared experiences

## Potential Extensions

### Dynamic World Generation
- Use post content to generate game world elements
- Create locations mentioned in posts as visitable areas

### Timeline Visualization
- Show character development over time based on post history
- Create "memory" system where NPCs recall past events

### Customizable Characters
- Allow minor customization of NPC appearance
- Generate character outfits based on activities mentioned in posts

### Minigames
- Create simple minigames based on activities mentioned in posts
- Allow players to participate in events described in blog posts

This implementation plan provides a comprehensive framework for integrating blog content with the Godot game, creating a dynamic and personalized browser gaming experience that reflects the real-world activities of blog authors.