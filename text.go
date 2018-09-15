package randtxt

import (
	"bytes"
	"errors"
	"io"
	"math/rand"
	"strings"
	"unicode"

	"github.com/pboyd/markov"
)

type Generator struct {
	chain     markov.Chain
	ngramSize int
}

func NewGenerator(chain markov.Chain, ngramSize int) *Generator {
	if ngramSize <= 1 {
		panic("ngramSize must be greater than 1")
	}

	return &Generator{
		chain:     chain,
		ngramSize: ngramSize,
	}
}

// Paragraph returns a paragraph containing between "min" and "max" sentences.
func (g *Generator) Paragraph(min, max int) (string, error) {
	total := rand.Intn(max-min) + min
	generated := 0

	done := make(chan struct{})
	defer close(done)

	gen := g.generate(done)
	text := &bytes.Buffer{}

	for te := range gen {
		if te.Tag.Tag == "." {
			break
		}
	}

	first := <-gen
	if first.Err != nil {
		return "", first.Err
	}
	g.writeWord(text, tag{}, first.Tag)

	last := first.Tag

	for te := range gen {
		if te.Err != nil {
			return "", te.Err
		}

		tag := te.Tag

		g.writeWord(text, last, tag)

		if tag.Tag == "." {
			generated++
			if generated == total {
				break
			}
		}

		last = tag
	}

	return text.String(), nil
}

func (g *Generator) generate(done chan struct{}) <-chan tagError {
	out := make(chan tagError)

	go func() {
		defer close(out)

		past, err := g.seed()
		if err != nil {
			select {
			case out <- tagError{Err: err}:
			case <-done:
			}

			return
		}

		for _, rawTag := range past {
			select {
			case out <- tagError{Tag: parseTag(rawTag)}:
			case <-done:
				return
			}
		}

		for {
			next, err := g.next(past)
			if err != nil {
				select {
				case out <- tagError{Err: err}:
				case <-done:
				}

				return
			}

			// Shift the past elements to the left to make room for
			// the new word.
			copy(past, past[1:g.ngramSize])
			past[g.ngramSize-1] = next

			select {
			case out <- tagError{Tag: parseTag(next)}:
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

func (g *Generator) writeWord(w io.Writer, prev, this tag) {
	needSpace := false

	switch this.Tag {
	case ".", ",", ":", "POS":
	case "RB":
		if this.Text != "n't" {
			needSpace = true
		}
	case "VBZ":
		if !strings.HasPrefix(this.Text, "'") {
			needSpace = true
		}
	default:
		needSpace = true
	}

	if !prev.IsZero() && needSpace {
		io.WriteString(w, " ")
	}

	word := this.Text

	switch prev.Tag {
	case "", ".", ":":
		if prev.Text != ";" {
			word = titleCase(word)
		}
	}

	io.WriteString(w, word)
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
func titleCase(text string) string {
	buf := []rune(text)
	if len(buf) == 0 {
		return ""
	}

	buf[0] = unicode.ToTitle(buf[0])
	return string(buf)
}

type tag struct {
	Text string
	Tag  string
}

func (t tag) IsZero() bool {
	return t.Text == "" && t.Tag == ""
}

func parseTag(gram string) tag {
	split := strings.Split(gram, "/")
	if len(split) < 2 {
		return tag{}
	}

	return tag{
		Text: split[0],
		Tag:  split[1],
	}
}

type tagError struct {
	Tag tag
	Err error
}
