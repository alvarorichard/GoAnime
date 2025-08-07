# AllAnime Episode Navigation Enhancement

This document describes the enhanced episode navigation features implemented for AllAnime sources, inspired by the ani-cli project.

## Features Implemented

### 1. AllAnime Navigator (`AllAnimeNavigator`)
A dedicated navigator class that handles AllAnime-specific episode navigation:

- **Next Episode Navigation**: Seamlessly moves to the next available episode
- **Previous Episode Navigation**: Moves to the previous episode
- **Episode Selection**: Direct episode selection by number
- **Quality Management**: Change video quality (best, 1080p, 720p, 480p, 360p, worst)
- **Mode Management**: Switch between subtitled (sub) and dubbed (dub) versions
- **Episode Listing**: Get complete list of available episodes

### 2. Enhanced Playback Controls
Updated the playback interface to include AllAnime-specific options:

#### Standard Options (All Sources)
- Next episode (`n`)
- Previous episode (`p`) 
- Replay current episode (`r`)
- Select episode (`e`)
- Change anime (`c`)
- Quit (`quit`)

#### AllAnime-Specific Options
- Change quality (`q`) - Allows selection of video quality
- Change mode (`m`) - Switch between sub/dub versions

### 3. Intelligent Source Detection
The system automatically detects AllAnime sources through multiple methods:

1. **Source Field**: Checks `anime.Source == "AllAnime"`
2. **Name Tags**: Detects `[AllAnime]` tags in anime names
3. **URL Analysis**: Identifies AllAnime URLs containing "allanime"
4. **ID Pattern**: Recognizes short alphanumeric AllAnime IDs

### 4. Enhanced API Integration
New API functions for AllAnime navigation:

- `GetEpisodeStreamURLEnhanced()`: Enhanced episode URL fetching
- `GetAllAnimeEpisodeWithNavigation()`: Navigation-aware episode retrieval
- `GetAllAnimeEpisodeList()`: Formatted episode list retrieval

## Implementation Details

### Key Files Modified/Created

1. **`internal/playback/allanime_navigation.go`** (NEW)
   - Contains the `AllAnimeNavigator` class
   - Implements episode navigation logic
   - Handles quality and mode management

2. **`internal/playback/series.go`** (ENHANCED)
   - Updated `handleUserNavigation()` to support AllAnime
   - Added `handleUserNavigationEnhanced()` function
   - Integrated AllAnime-specific menu options

3. **`internal/playback/input.go`** (ENHANCED)
   - Added `GetUserInputEnhanced()` function
   - Provides context-aware menu options
   - AllAnime sources get additional quality/mode options

4. **`internal/api/allanime_enhanced.go`** (NEW)
   - Enhanced API functions for AllAnime
   - Navigation-aware episode URL fetching
   - Source detection and validation

5. **`internal/player/scraper.go`** (ENHANCED)
   - Updated `GetVideoURLForEpisodeEnhanced()` function
   - Integrated AllAnime navigation support
   - Fallback mechanism for non-AllAnime sources

### Navigation Flow

```
User Input → GetUserInputEnhanced() → handleUserNavigationEnhanced()
    ↓
AllAnime Source? → YES → handleAllAnimeNavigation()
    ↓                     ↓
    NO                AllAnimeNavigator
    ↓                     ↓
Regular Navigation    Enhanced Navigation
    ↓                     ↓
Standard Episode      Next/Previous with
Selection            AllAnime API
```

## ani-cli Compatibility

This implementation closely follows the ani-cli navigation model:

### Similar Features
- **Next/Previous**: Direct episode navigation like ani-cli's `next`/`previous` commands
- **Quality Selection**: Multiple quality options similar to ani-cli's quality system
- **Episode Selection**: Interactive episode picker
- **Replay**: Ability to replay current episode
- **Mode Switching**: Sub/dub switching (AllAnime specific)

### Enhanced Features Beyond ani-cli
- **Source-Aware Navigation**: Automatic detection and handling of different sources
- **Quality Management**: Interactive quality selection menu
- **Mode Management**: Interactive sub/dub switching
- **Fallback Support**: Graceful fallback to regular navigation for non-AllAnime sources

## Usage Examples

### Basic Navigation
```go
// Next episode
selectedEpisode, err := HandleAllAnimeEpisodeNavigation(anime, episodes, currentEpisode, "next")

// Previous episode  
selectedEpisode, err := HandleAllAnimeEpisodeNavigation(anime, episodes, currentEpisode, "previous")
```

### Navigator Usage
```go
navigator, err := NewAllAnimeNavigator(anime)
if err != nil {
    // Handle error
}

// Get next episode
nextEp, err := navigator.GetNextEpisode(currentEpisode)

// Change quality
navigator.SetQuality("720p")

// Change mode
err = navigator.SetMode("dub")
```

### Enhanced API Usage
```go
// Get episode URL with navigation
episode, streamURL, err := GetAllAnimeEpisodeWithNavigation(anime, "5", "next")

// Get enhanced stream URL
streamURL, err := GetEpisodeStreamURLEnhanced(episode, anime, "best")
```

## Error Handling

The implementation includes comprehensive error handling:

- **Invalid Navigation**: Prevents navigation beyond first/last episodes
- **Source Validation**: Ensures AllAnime functions only work with AllAnime sources
- **API Failures**: Graceful fallback to regular navigation on API errors
- **Quality/Mode Errors**: User-friendly error messages for invalid selections

## Testing

Unit tests are provided in `allanime_navigation_test.go`:

- Source detection validation
- ID extraction testing
- Navigator creation testing
- Navigation command testing

## Future Enhancements

Potential improvements based on ani-cli features:

1. **Skip Intro/Outro**: Integration with AniSkip for automatic skipping
2. **Playlist Support**: Batch episode management
3. **Download Integration**: Enhanced download workflow for AllAnime
4. **Subtitle Management**: Better subtitle handling
5. **Search Integration**: Direct anime search from navigation

## Compatibility Notes

- **Backward Compatibility**: All existing functionality remains unchanged
- **Source Agnostic**: Non-AllAnime sources continue to work normally
- **Optional Features**: AllAnime-specific features only appear for AllAnime sources
- **Graceful Degradation**: Falls back to regular navigation on errors

This implementation provides a superior episode navigation experience for AllAnime sources while maintaining full compatibility with existing animefire.plus functionality.
