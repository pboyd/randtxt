package randtxt

import (
	"testing"

	"github.com/pboyd/markov"
)

func TestBuilder(t *testing.T) {
	chain := markov.NewMemoryChain(100)
	b := NewModelBuilder(chain, 3)

	err := b.Feed(tagFeed(100))
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	root, err := chain.Get(0)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	expectedRoot := "A/NNP B/NNP C/NNP"
	if root.(string) != expectedRoot {
		t.Errorf("got %q, want %q", root, expectedRoot)
	}
}

func tagFeed(n int) <-chan Tag {
	c := make(chan Tag)

	go func() {
		defer close(c)

		letters := []string{"A", "B", "C", "D", "E", "F", "G"}
		for i := 0; i < n; i++ {
			c <- Tag{
				Text: letters[i%len(letters)],
				POS:  "NNP",
			}
		}
	}()

	return c
}
