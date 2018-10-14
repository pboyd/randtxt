package randtxt

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/pboyd/markov"
)

// Generator generates random text from a model built by ModelBuilder.
type Generator struct {
	chain     markov.Chain
	ngramSize int

	// TagSet is the language and tagset specific rules. This should match
	// the TagSet used when the model was built.
	TagSet TagSet
}

// NewGenerator returns a new generator. Returns an error if the chain has an
// unrecognized format.
func NewGenerator(chain markov.Chain) (*Generator, error) {
	ngramSize, err := inspectChain(chain)
	if err != nil {
		return nil, err
	}

	return &Generator{
		chain:     chain,
		ngramSize: ngramSize,
		TagSet:    PennTreebankTagSet,
	}, nil
}

func inspectChain(chain markov.Chain) (int, error) {
	root, err := chain.Get(0)
	if err != nil {
		return 0, err
	}

	rootString, ok := root.(string)
	if !ok {
		return 0, fmt.Errorf("chain has type %T, want string", root)
	}

	split := strings.Split(rootString, " ")
	if len(split) < 2 {
		return 0, fmt.Errorf("chain has ngram size of %d, want at least 2", len(split))
	}

	for _, gram := range split {
		tag := parseTag(gram)
		if tag.POS == "" || tag.Text == "" {
			return 0, fmt.Errorf("unrecognized tag format %q", gram)
		}
	}

	return len(split), nil
}

// Paragraph returns a paragraph containing between "min" and "max" sentences.
func (g *Generator) Paragraph(min, max int) (string, error) {
	text := &bytes.Buffer{}
	err := g.WriteParagraph(text, min, max)
	if err != nil {
		return "", err
	}

	return text.String(), nil
}

// WriteParagraph writes a paragraph of random text to "out". The paragraph
// will contain between "min" and "max" sentences.
func (g *Generator) WriteParagraph(out io.Writer, min, max int) error {
	total := rand.Intn(max-min) + min
	generated := 0

	done := make(chan struct{})
	defer close(done)

	gen := g.generate(done)

	for te := range gen {
		if te.Tag.POS == "." {
			break
		}
	}

	first := <-gen
	if first.Err != nil {
		return first.Err
	}
	io.WriteString(out, g.TagSet.Join(first.Tag, Tag{}))

	last := first.Tag

	for te := range gen {
		if te.Err != nil {
			return te.Err
		}

		tag := te.Tag

		io.WriteString(out, g.TagSet.Join(tag, last))

		if tag.POS == "." {
			generated++
			if generated == total {
				break
			}
		}

		last = tag
	}

	return nil
}

func (g *Generator) generate(done chan struct{}) <-chan tagOrError {
	out := make(chan tagOrError)

	go func() {
		defer close(out)

		past, err := g.seed()
		if err != nil {
			select {
			case out <- tagOrError{Err: err}:
			case <-done:
			}

			return
		}

		for _, rawTag := range past {
			select {
			case out <- tagOrError{Tag: parseTag(rawTag)}:
			case <-done:
				return
			}
		}

		for {
			next, err := g.next(past)
			if err != nil {
				select {
				case out <- tagOrError{Err: err}:
				case <-done:
				}

				return
			}

			// Shift the past elements to the left to make room for
			// the new word.
			copy(past, past[1:g.ngramSize])
			past[g.ngramSize-1] = next

			select {
			case out <- tagOrError{Tag: parseTag(next)}:
			case <-done:
				return
			}
		}
	}()

	return out
}

func (g *Generator) seed() ([]string, error) {
	for {
		raw, err := markov.Random(g.chain)
		if err != nil {
			return nil, err
		}

		seed := strings.Split(raw.(string), " ")
		if len(seed) > 1 {
			return seed, nil
		}
	}
}

func (g *Generator) next(past []string) (string, error) {
	key := strings.Join(past, " ")
	pastID, err := g.chain.Find(key)
	if err != nil {
		return "", err
	}

	links, err := g.chain.Links(pastID)
	if err != nil {
		return "", err
	}

	if len(links) == 0 {
		return "", errors.New("not found")
	}

	index := rand.Float64()
	var passed float64

	for _, link := range links {
		passed += link.Probability
		if passed > index {
			raw, err := g.chain.Get(link.ID)
			if err != nil {
				return "", err
			}

			return raw.(string), nil
		}
	}

	return "", errors.New("failed")
}

type tagOrError struct {
	Tag Tag
	Err error
}
