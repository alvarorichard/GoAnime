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
	Source    string    // Identifies the source (AllAnime, AnimeFire, FlixHQ, etc.)
	MediaType MediaType // Type of media (anime, movie, tv)
	Year      string    // Release year
	Quality   string    // Video quality (if available)
	// TMDB fields for movies/TV shows
	TMDBID      int          // TMDB ID
	IMDBID      string       // IMDB ID
	TMDBDetails *TMDBDetails // Detailed TMDB information
	Rating      float64      // Rating (0-10)
	Overview    string       // Description/synopsis
	Genres      []string     // Genre list
	Runtime     int          // Runtime in minutes (for movies)
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

// TMDBDetails contains movie/TV show information from TMDB
type TMDBDetails struct {
	ID                  int            `json:"id"`
	IMDBID              string         `json:"imdb_id"`
	Title               string         `json:"title"` // For movies
	Name                string         `json:"name"`  // For TV shows
	OriginalName        string         `json:"original_name"`
	Overview            string         `json:"overview"`
	Tagline             string         `json:"tagline"`
	PosterPath          string         `json:"poster_path"`
	BackdropPath        string         `json:"backdrop_path"`
	ReleaseDate         string         `json:"release_date"`   // For movies
	FirstAirDate        string         `json:"first_air_date"` // For TV shows
	VoteAverage         float64        `json:"vote_average"`
	VoteCount           int            `json:"vote_count"`
	Popularity          float64        `json:"popularity"`
	Runtime             int            `json:"runtime"` // For movies (minutes)
	Status              string         `json:"status"`
	Genres              []TMDBGenre    `json:"genres"`
	ProductionCompanies []TMDBCompany  `json:"production_companies"`
	SpokenLanguages     []TMDBLanguage `json:"spoken_languages"`
	NumberOfSeasons     int            `json:"number_of_seasons"`  // For TV shows
	NumberOfEpisodes    int            `json:"number_of_episodes"` // For TV shows
}

// TMDBGenre represents a genre from TMDB
type TMDBGenre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TMDBCompany represents a production company from TMDB
type TMDBCompany struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	LogoPath string `json:"logo_path"`
	Country  string `json:"origin_country"`
}

// TMDBLanguage represents a spoken language from TMDB
type TMDBLanguage struct {
	ISO639      string `json:"iso_639_1"`
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
}

// TMDBSearchResult represents a search result from TMDB
type TMDBSearchResult struct {
	Page         int         `json:"page"`
	TotalResults int         `json:"total_results"`
	TotalPages   int         `json:"total_pages"`
	Results      []TMDBMedia `json:"results"`
}

// TMDBMedia represents a movie or TV show in search results
type TMDBMedia struct {
	ID            int     `json:"id"`
	MediaType     string  `json:"media_type"` // "movie" or "tv"
	Title         string  `json:"title"`      // For movies
	Name          string  `json:"name"`       // For TV shows
	OriginalTitle string  `json:"original_title"`
	OriginalName  string  `json:"original_name"`
	Overview      string  `json:"overview"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	ReleaseDate   string  `json:"release_date"`
	FirstAirDate  string  `json:"first_air_date"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	Adult         bool    `json:"adult"`
	GenreIDs      []int   `json:"genre_ids"`
}

// GetDisplayTitle returns the appropriate title for the media
func (m *TMDBMedia) GetDisplayTitle() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

// GetReleaseYear returns the release year
func (m *TMDBMedia) GetReleaseYear() string {
	date := m.ReleaseDate
	if date == "" {
		date = m.FirstAirDate
	}
	if len(date) >= 4 {
		return date[:4]
	}
	return ""
}

// GetPosterURL returns the full poster URL
func (m *TMDBMedia) GetPosterURL(size string) string {
	if m.PosterPath == "" {
		return ""
	}
	if size == "" {
		size = "w500"
	}
	return "https://image.tmdb.org/t/p/" + size + m.PosterPath
}

// GetBackdropURL returns the full backdrop URL
func (m *TMDBMedia) GetBackdropURL(size string) string {
	if m.BackdropPath == "" {
		return ""
	}
	if size == "" {
		size = "w1280"
	}
	return "https://image.tmdb.org/t/p/" + size + m.BackdropPath
}

// TMDBCredits contains cast and crew information
type TMDBCredits struct {
	ID   int        `json:"id"`
	Cast []TMDBCast `json:"cast"`
	Crew []TMDBCrew `json:"crew"`
}

// TMDBCast represents a cast member
type TMDBCast struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
}

// TMDBCrew represents a crew member
type TMDBCrew struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Department  string `json:"department"`
	Job         string `json:"job"`
	ProfilePath string `json:"profile_path"`
}

// TMDBSeason represents a TV show season from TMDB
type TMDBSeason struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	AirDate      string `json:"air_date"`
	PosterPath   string `json:"poster_path"`
}

// TMDBEpisode represents a TV episode from TMDB
type TMDBEpisode struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Overview      string  `json:"overview"`
	EpisodeNumber int     `json:"episode_number"`
	SeasonNumber  int     `json:"season_number"`
	AirDate       string  `json:"air_date"`
	Runtime       int     `json:"runtime"`
	StillPath     string  `json:"still_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
}
