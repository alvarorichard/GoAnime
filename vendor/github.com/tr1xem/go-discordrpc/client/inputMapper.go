package client

import (
	"time"
)

// Activity holds the data for discord rich presence
type Activity struct {
	// Activity Type; 0=Playing, 1=Streaming, 2=Listening, 3=Watching
	Type int
	// What the player is currently doing
	Name string
	// What the player is currently doing
	Details string
	// The user's current party status
	State string
	// The id for a large asset of the activity, usually a snowflake
	LargeImage string
	// Text displayed when hovering over the large image of the activity
	LargeText string
	// The id for a small asset of the activity, usually a snowflake
	SmallImage string
	// Text displayed when hovering over the small image of the activity
	SmallText string
	// Information for the current party of the player
	Party *Party
	// Unix timestamps for start and/or end of the game
	Timestamps *Timestamps
	// Secrets for Rich Presence joining and spectating
	Secrets *Secrets
	// Clickable buttons that open a URL in the browser
	Buttons []*Button
}

// Button holds a label and the corresponding URL that is opened on press
type Button struct {
	// The label of the button
	Label string
	// The URL of the button
	Url string
}

// Party holds information for the current party of the player
type Party struct {
	// The ID of the party
	ID string
	// Used to show the party's current size
	Players int
	// Used to show the party's maximum size
	MaxPlayers int
}

// Timestamps holds unix timestamps for start and/or end of the game
type Timestamps struct {
	// unix time (in milliseconds) of when the activity started
	Start *time.Time
	// unix time (in milliseconds) of when the activity ends
	End *time.Time
}

// Secrets holds secrets for Rich Presence joining and spectating
type Secrets struct {
	// The secret for a specific instanced match
	Match string
	// The secret for joining a party
	Join string
	// The secret for spectating a game
	Spectate string
}

func mapActivity(activity *Activity) *PayloadActivity {
	final := &PayloadActivity{
		Name:    activity.Name,
		Details: activity.Details,
		State:   activity.State,
		Assets: PayloadAssets{
			LargeImage: activity.LargeImage,
			LargeText:  activity.LargeText,
			SmallImage: activity.SmallImage,
			SmallText:  activity.SmallText,
		},
		Type: activity.Type,
	}

	if activity.Timestamps != nil && activity.Timestamps.Start != nil {
		start := uint64(activity.Timestamps.Start.UnixNano() / 1e6)
		final.Timestamps = &PayloadTimestamps{
			Start: &start,
		}
		if activity.Timestamps.End != nil {
			end := uint64(activity.Timestamps.End.UnixNano() / 1e6)
			final.Timestamps.End = &end
		}
	}

	if activity.Party != nil {
		final.Party = &PayloadParty{
			ID:   activity.Party.ID,
			Size: [2]int{activity.Party.Players, activity.Party.MaxPlayers},
		}
	}

	if activity.Secrets != nil {
		final.Secrets = &PayloadSecrets{
			Join:     activity.Secrets.Join,
			Match:    activity.Secrets.Match,
			Spectate: activity.Secrets.Spectate,
		}
	}

	if len(activity.Buttons) > 0 {
		for _, btn := range activity.Buttons {
			final.Buttons = append(final.Buttons, &PayloadButton{
				Label: btn.Label,
				Url:   btn.Url,
			})
		}
	}

	return final
}
