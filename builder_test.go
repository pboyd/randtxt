package randtxt

import (
	"fmt"
	"sort"
	"strings"
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

	actual := describeID(chain, 0)
	expected := "A/NNP B/NNP C/NNP:D/NNP-1.00"
	if actual != expected {
		t.Errorf("got %q, want %q", actual, expected)
	}
}

func TestUnigramBuilder(t *testing.T) {
	chain := markov.NewMemoryChain(100)
	b := NewModelBuilder(chain, 1)

	err := b.Feed(tagFeed(100))
	if err != nil {
		t.Fatalf("got error: %v", err)
	}

	actual := describeID(chain, 0)
	expected := "A/NNP:B/NNP-1.00"
	if actual != expected {
		t.Errorf("got %q, want %q", actual, expected)
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

func describeID(chain markov.Chain, id int) string {
	links, err := chain.Links(0)
	if err != nil {
		return err.Error()
	}

	linkDesc := make([]string, len(links))

	for i, link := range links {
		val, err := chain.Get(link.ID)
		if err != nil {

		}
		linkDesc[i] = fmt.Sprintf("%s-%0.2f", val, link.Probability)
	}

	sort.Strings(linkDesc)

	value, err := chain.Get(0)
	if err != nil {
		return err.Error()
	}

	return value.(string) + ":" + strings.Join(linkDesc, " ")
}
