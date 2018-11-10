package randtxt

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"github.com/pboyd/markov"
)

// Generator generates random text from a model built by ModelBuilder.
type Generator struct {
	chain markov.Chain

	// TagSet is the language and tagset specific rules. This should match
	// the TagSet used when the model was built.
	TagSet TagSet
}

// NewGenerator returns a new generator. Returns an error if the chain has an
// unrecognized format.
func NewGenerator(chain markov.Chain) (*Generator, error) {
	_, err := inspectChain(chain)
	if err != nil {
		return nil, err
	}

	return &Generator{
		chain:  chain,
		TagSet: PennTreebankTagSet,
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

	send := func(tag Tag, err error) bool {
		te := tagOrError{
			Tag: tag,
			Err: err,
		}

		select {
		case out <- te:
			return false
		case <-done:
			return true
		}
	}

	go func() {
		defer close(out)

		past, err := randomSeed(g.chain)
		if err != nil {
			send(Tag{}, err)
			return
		}

		for _, rawTag := range strings.Split(past, " ") {
			if send(parseTag(rawTag), nil) {
				return
			}
		}

		model, err := NewModel(g.chain, past)
		if err != nil {
			send(Tag{}, err)
			return
		}

		for {
			err := model.Step()
			if err != nil {
				send(Tag{}, err)
				return
			}

			if send(model.Current(), nil) {
				return
			}
		}
	}()

	return out
}

type tagOrError struct {
	Tag Tag
	Err error
}
