package test_util_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

func TestTreatingAnimeName(t *testing.T) {
	t.Run("Should generate an lowercase slug with only spaces replaced to dashes", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"Naruto Shippuden", "naruto-shippuden"},
			{"My Hero Academia", "my-hero-academia"},
			{"Attack on Titan", "attack-on-titan"},
			{"One Piece", "one-piece"},
			{"Sword Art Online", "sword-art-online"},
			{"Fullmetal Alchemist: Brotherhood", "fullmetal-alchemist:-brotherhood"},
			{"Death Note", "death-note"},
			{"Dragon Ball Z", "dragon-ball-z"},
			{"Tokyo Ghoul", "tokyo-ghoul"},
			{"Demon Slayer: Kimetsu no Yaiba", "demon-slayer:-kimetsu-no-yaiba"},
			{"Re:ZERO -Starting Life in Another World-", "re:zero--starting-life-in-another-world-"},
			{"Steins;Gate", "steins;gate"},
			{"Kaguya-sama: Love is War", "kaguya-sama:-love-is-war"},
			{"Made in Abyss", "made-in-abyss"},
			{"No Game No Life", "no-game-no-life"},
			{"The Promised Neverland", "the-promised-neverland"},
			{"One Punch Man", "one-punch-man"},
			{"Hunter x Hunter", "hunter-x-hunter"},
			{"Your Lie in April", "your-lie-in-april"},
			{"Sword Art Online II", "sword-art-online-ii"},
		}

		for _, test := range tests {
			output := util.TreatingAnimeName(test.input)
			if output != test.expected {
				t.Errorf("For input '%s', expected '%s', but got '%s'", test.input, test.expected, output)
			}
		}
	})
}
