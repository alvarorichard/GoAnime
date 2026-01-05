package models

type Anime struct {
	Name      string
	URL       string
	ImageURL  string
	Episodes  []Episode
	AnilistID int
	MalID     int
	Details   AniListDetails
	Source    string // Identifies the source (AllAnime, AnimeFire, etc.)
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
