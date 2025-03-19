# Game Integration System Architecture

```mermaid
graph TD
    %% GitHub Components
    GH[GitHub Repository: www-jp] --> GHI[GitHub Issues]
    GHI --> |labeled as 'post'| PC[Post Comment '/post']
    GHR[GitHub Repository: game] --> |stores game code| WASM[WASM Build]
    
    %% GitHub Actions
    subgraph "GitHub Actions in www-jp repo"
        PW[Post Workflow] --> |triggered by '/post'| PP[Post Processor]
        PP --> |extracts post raw content| GDP[Game Data Processor]
        PP --> |processes post content| TP[Translator Process]
        GDP --> |parallel process| GDG[Game Data Generator]
        TP --> |translates to Japanese| D[Deploy Site]
        GDG --> |generates character data| VS[Voice Synthesis]
        VS --> |creates audio files| DSU[Data Storage Updater]
        DSU --> |pushes game assets| GCR[Game Content Repository Update]
    end
    
    subgraph "GitHub Actions in game repo"
        GCT[Game Content Trigger] --> |triggered by commit| GB[Game Build]
        GB --> |builds WASM version| GD[Game Deployment]
    end
    
    %% External Services
    subgraph "AI Services"
        OAI[OpenAI API] --- |translates content| TP
        OAI --- |generates structured data| GDG
        EL[ElevenLabs API] --- |generates voice clips| VS
    end
    
    %% Storage Services
    subgraph "Storage Services"
        GCS[Google Cloud Storage] --- |stores audio assets| DSU
        GCS --- |lazy loads audio| WASM
    end
    
    %% Game Components
    subgraph "Game System (Godot)"
        WASM[WASM Game Build] --> |runs in browser| WSB[Web Browser]
        WASM --> |loads data from| GHR
        WASM --> |lazy loads audio from| GCS
    end
    
    %% Data Flows
    PC --> PW
    GCR --> GCT
    
    %% Game Player Interaction
    WSB --> |presents to| P[Player]
    P --> |interacts with| NPC[NPC Characters]
    NPC --> |displays dialogue from| WASM
    NPC --> |plays audio from| GCS
    
    %% Visual styling
    classDef github fill:#f9f9f9,stroke:#333,stroke-width:1px
    classDef actions fill:#e1f5fe,stroke:#01579b,stroke-width:1px
    classDef services fill:#e8f5e9,stroke:#2e7d32,stroke-width:1px
    classDef storage fill:#fff3e0,stroke:#e65100,stroke-width:1px
    classDef game fill:#f3e5f5,stroke:#6a1b9a,stroke-width:1px
    classDef user fill:#fce4ec,stroke:#880e4f,stroke-width:1px
    
    class GH,GHI,PC,GHR github
    class PW,PP,TP,D,GDG,VS,DSU,GCR,GCT,GB,GD,GDP actions
    class OAI,EL services
    class GCS storage
    class WASM,WSB,NPC game
    class P user
```

## Component Descriptions

### GitHub Components
- **GitHub Repository: www-jp**: Blog content repo with posts
- **GitHub Issues**: Used to create blog post content with "post" label
- **Post Comment**: Trigger ("/post") to initiate the workflow
- **GitHub Repository: game**: Separate repo containing Godot game code

### GitHub Actions (Workflows)
#### In www-jp repo:
- **Post Workflow**: Main workflow triggered by "/post" comment
- **Post Processor**: Processes issue content
- **Game Data Processor**: NEW - Extracts raw post data early in the process
- **Translator Process**: Translates content to Japanese using OpenAI
- **Deploy Site**: Deploys the updated blog with new content
- **Game Data Generator**: NEW - Converts raw post data to structured game data
- **Voice Synthesis**: NEW - Creates audio files from dialogue text
- **Data Storage Updater**: NEW - Updates GCS with audio assets
- **Game Content Repository Update**: NEW - Pushes game data to game repo

#### In game repo:
- **Game Content Trigger**: Triggered by content repo updates
- **Game Build**: Builds the Godot game to WASM
- **Game Deployment**: Deploys WASM build to web hosting

### AI Services
- **OpenAI API**: Used for translation and structured data generation
- **ElevenLabs API**: Used for voice synthesis from text

### Storage Services
- **Google Cloud Storage**: Stores generated audio assets for lazy loading

### Game System
- **WASM Game Build**: Browser-compatible build of the Godot game
- **Web Browser**: Platform where the game runs
- **NPC Characters**: In-game representations of blog authors

### User Interaction
- **Player**: End user interacting with the game
- **NPC Interaction**: Dialogue and behavior systems in the game

## Data Flow Sequence

1. Blog post is created via GitHub Issue with "post" label
2. "/post" comment triggers Post Workflow
3. Post Processor extracts raw post content
4. Game Data Processor takes the raw content (parallel to translation)
5. Game Data Generator creates structured JSON data for game:
   - Character traits and movement patterns
   - Dialogue options categorized by topic
   - Activity records
6. Voice Synthesis converts dialogue text to audio in both languages
7. Data Storage Updater:
   - Uploads audio files to Google Cloud Storage
   - Maintains references between text and audio assets
8. Game Content Repository Update:
   - Pushes JSON data and asset references to game repo
   - Triggers a webhook or workflow dispatch to game repo
9. Game Build process in game repo:
   - Triggered by content update
   - Builds Godot project to WASM 
   - Deploys to web hosting
10. Game in browser:
    - Loads character data directly from game code
    - Lazy loads audio assets from GCS as needed
11. Player interacts with NPCs representing blog authors

## Integration Points

### www-jp repo → game repo
- Structured JSON data for character attributes and dialogue
- Asset references for audio files in GCS
- Webhook or workflow_dispatch trigger for build

### GitHub Actions → AI Services
- Structured API calls with appropriate context and prompts
- Authentication using API keys stored as GitHub Secrets

### GitHub Actions → GCS
- Direct upload of audio files to Cloud Storage
- Generation of public URLs for game access

### Game → GCS
- Lazy loading of audio assets during gameplay
- Caching mechanisms for frequently accessed assets

### Game → Player
- Character behavior systems driven by generated data
- Dialogue UI with language selection
- Audio playback for immersive experience