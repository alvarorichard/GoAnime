// Package models contains TMDB (The Movie Database) data structures
package models

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
