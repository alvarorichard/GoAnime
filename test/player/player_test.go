package player_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/player"
)

// TODO: Find out why Mugyutto! Black Clover episode 3 is failing...

func TestGetVideoURLForEpisode(t *testing.T) {
	type test_case struct {
		baseUrl string
		expect  string
	}

	testCases := []test_case{
		{baseUrl: "https://animefire.plus/animes/new-game/1", expect: "https://lightspeedst.net/s3/mp4/new-game/sd/1.mp4"},
		{baseUrl: "https://animefire.plus/animes/new-game/2", expect: "https://lightspeedst.net/s3/mp4/new-game/sd/2.mp4"},
		{baseUrl: "https://animefire.plus/animes/new-game/3", expect: "https://lightspeedst.net/s3/mp4/new-game/sd/3.mp4"},
		{baseUrl: "https://animefire.plus/animes/mugyutto-black-clover/1", expect: "https://www.blogger.com/video.g?token=AD6v5dwHn9BT-mLhC890nIu9lom06mPy4S9P5QWUfRwQnJe3J8DOkMHvRGJyRV27r5G2pAAQQxrTukxPkRNUaFiq4JJUyRM1yorh9iqmO5soutymtm3AhdlYII4NJ6T69_G3ALs8w1Y"},
		{baseUrl: "https://animefire.plus/animes/mugyutto-black-clover/2", expect: "https://www.blogger.com/video.g?token=AD6v5dzO93oJWeKZUsAlYh8b4lsxXkiaJv8yzDsO3ABhBEvVx7ctUlbj0VVSiAmApFdKvMLQVHQ2Y7DNuOnup8EBM40Z71aBW28FtDAqs52piuiZbCHecpa-3iddlhKR5yvCHLW18Gc"},
		{baseUrl: "https://animefire.plus/animes/mugyutto-black-clover/3", expect: "https://www.blogger.com/video.g?token=AD6v5dwsyYNo5Xf1d-27hbuzblqfTt4iuUKfLwVXroMkZrv3JOgixG2KuMt8lfobrM15mx0QNq8kyKggd2775PczW-2dlEwfoHb7wnk3eVeqQXyiOGmVWbaQyVOeJgfteRGnijX8F20O"},
	}

	for _, tc := range testCases {
		t.Run(tc.baseUrl, func(t *testing.T) {
			got, err := player.GetVideoURLForEpisode(tc.baseUrl)
			if err != nil {
				t.Fatalf("GetVideoURLForEpisode() error = %v", err)
				return
			}
			if got != tc.expect {
				t.Errorf("GetVideoURLForEpisode() = %v, want %v", got, tc.expect)
			}
		})
	}
}
