package randtxt

import (
	"os"
	"regexp"
	"testing"

	"github.com/pboyd/markov"
)

func TestParagraph(t *testing.T) {
	chain, close := testChain(t, "testfiles/ion/trigram.mkv")
	defer close()

	const (
		min = 3
		max = 5
	)

	g, err := NewGenerator(chain)
	if err != nil {
		t.Fatalf("invalid chain: %v", err)
	}

	text, err := g.Paragraph(3, 5)
	if err != nil {
		t.Fatalf("got error %v, want nil", err)
	}

	matchTag := regexp.MustCompile(`/[A-Z$.:]+`)
	if matchTag.MatchString(text) {
		t.Errorf("text contained POS tags")
	}

	matchSpaceBeforePunctuation := regexp.MustCompile(` [,.?!:]`)
	if matchSpaceBeforePunctuation.MatchString(text) {
		t.Errorf("text contained spaces before punctuation")
	}

	matchLowerCaseSentenceStart := regexp.MustCompile(`[.?!] [a-z]`)
	if matchLowerCaseSentenceStart.MatchString(text) {
		t.Errorf("text contained lower case letters at the beginning of a sentence")
	}

	sentenceEndings := regexp.MustCompile(`[.?!]`).FindAllString(text, -1)
	if len(sentenceEndings) < min {
		t.Errorf("got %d sentences, want at least %d", len(sentenceEndings), min)
	}

	if len(sentenceEndings) > max {
		t.Errorf("got %d sentences, want at most %d", len(sentenceEndings), max)
	}

	matchFullSentence := regexp.MustCompile(`[.?!]$`)
	if !matchFullSentence.MatchString(text) {
		t.Errorf("text did not end with a sentence ending")
	}
}

func testChain(t *testing.T, path string) (chain markov.Chain, close func() error) {
	t.Helper()
	fh, err := os.Open(path)
	if err != nil {
		t.Fatalf("could not open %q: %v", path, err)
	}

	chain, err = markov.ReadDiskChain(fh)
	if err != nil {
		t.Fatalf("could not read chain from %q: %v", path, err)
	}

	return chain, fh.Close
}

func TestValidChain(t *testing.T) {
	chain, close := testChain(t, "testfiles/ion/trigram.mkv")
	defer close()

	_, err := NewGenerator(chain)
	if err != nil {
		t.Errorf("got error for valid chain: %v", err)
	}
}

func TestInvalidChains(t *testing.T) {
	cases := map[string]markov.Chain{
		"non-string":  invalidChain(1),
		"untagged":    invalidChain("Paul George Ringo"),
		"unknown tag": invalidChain("Paul;NN George;NN Ringo;NN"),
		"just tags":   invalidChain("/NN /VBZ"),
	}

	for desc, chain := range cases {
		_, err := NewGenerator(chain)
		if err == nil {
			t.Errorf("%s: no error", desc)
		}
	}
}

func invalidChain(root interface{}) markov.Chain {
	chain := markov.NewMemoryChain(1)
	chain.Add(root)
	return chain
}
