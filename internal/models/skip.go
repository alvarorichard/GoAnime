package models

// Skip represents a skip interval with a start and end time
type Skip struct {
	Start int
	End   int
}

// SkipTimes holds the skip intervals for OP and ED
type SkipTimes struct {
	Op Skip
	Ed Skip
}

// skipTimesResponse struct to hold the response from the AniSkip API
type SkipTimesResponse struct {
	Found   bool         `json:"found"`
	Results []SkipResult `json:"results"`
}

// SkipResult struct to hold individual skip result data
type SkipResult struct {
	Interval SkipInterval `json:"interval"`
	Type     string       `json:"skip_type"` // Corrected the tag to match the JSON response
}

// skipInterval struct to hold the start and end times for skip intervals
type SkipInterval struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}
