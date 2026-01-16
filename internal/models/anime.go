package models

// MediaType represents the type of media content
type MediaType string

const (
	MediaTypeAnime MediaType = "anime"
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

type Anime struct {
	Name      string
	URL       string
	ImageURL  string
	Episodes  []Episode
	AnilistID int
	MalID     int
	Details   AniListDetails
	Source    string // Identifies the source (AllAnime, AnimeFire, FlixHQ, etc.)
	MediaType MediaType // Type of media (anime, movie, tv)
	Year      string    // Release year
	Quality   string    // Video quality (if available)
}

// Media represents a movie or TV show (alias for better semantics)
type Media = Anime

// Season represents a TV show season
type Season struct {
	ID       string
	Number   int
	Title    string
	Episodes []Episode
}

// Episode represents a single episode of an anime series, containing details such as episode number,
// URL, title information, air date, duration, filler/recap status, synopsis, and skip times.
type Episode struct {
	Number    string
	Num       int
	URL       string
	Title     TitleDetails
	Aired     string
	Duration  int
	IsFiller  bool
	IsRecap   bool
	Synopsis  string
	SkipTimes SkipTimes
	DataID    string // Used for FlixHQ episode identification
	SeasonID  string // Season identifier for TV shows
}

// Subtitle represents a subtitle track for video playback
type Subtitle struct {
	URL      string
	Language string
	Label    string
	IsForced bool
}

// StreamInfo contains streaming information including video URL and subtitles
type StreamInfo struct {
	VideoURL   string
	Quality    string
	Subtitles  []Subtitle
	Referer    string
	SourceName string
	Headers    map[string]string
}

type TitleDetails struct {
	Romaji   string
	English  string
	Japanese string
}

type AniListResponse struct {
	Data struct {
		Media AniListDetails `json:"Media"`
	} `json:"data"`
}

type AniListDetails struct {
	ID           int         `json:"id"`
	IDMal        int         `json:"idMal"`
	Title        Title       `json:"title"`
	Description  string      `json:"description"`
	Genres       []string    `json:"genres"`
	AverageScore int         `json:"averageScore"`
	Episodes     int         `json:"episodes"`
	Status       string      `json:"status"`
	CoverImage   CoverImages `json:"coverImage"`
	Synonyms     []string    `json:"synonyms"`
}
type Title struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
}
type CoverImages struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}
