package randtxt

import (
	"errors"
	"math/rand"
	"strings"

	"github.com/pboyd/markov"
)

// Model steps through the chain word by word and reports probabilities.
//
// Model is a lower-level interface. Generator is the recommended way to
// generate text.
type Model struct {
	chain   markov.Chain
	current string
	past    []string
}

// NewModel initializes a model from a chain. "seed" is used as the starting
// point. If "seed" is blank a random seed is chosen.
func NewModel(chain markov.Chain, seed string) (*Model, error) {
	if seed == "" {
		var err error
		seed, err = randomSeed(chain)
		if err != nil {
			return nil, err
		}
	}

	// Make sure the seed exists.
	_, err := chain.Find(seed)
	if err != nil {
		return nil, err
	}

	past := strings.Split(seed, " ")
	if len(seed) == 1 {
		return nil, errors.New("invalid chain")
	}

	return &Model{
		chain: chain,
		past:  past,
	}, nil
}

func randomSeed(chain markov.Chain) (string, error) {
	for {
		raw, err := markov.Random(chain)
		if err != nil {
			return "", err
		}

		seed := strings.Split(raw.(string), " ")
		if len(seed) > 1 {
			return raw.(string), nil
		}
	}
}

// Current returns the word and POS tag that the model is currently at..
func (m *Model) Current() Tag {
	return parseTag(m.current)
}

// NextTags returns a list of tags that could be next along with their
// probabilities.
func (m *Model) NextTags() ([]TagProbability, error) {
	links, err := m.nextLinks()
	if err != nil {
		return nil, err
	}

	tp := make([]TagProbability, len(links))
	for i, link := range links {
		raw, err := m.chain.Get(link.ID)
		if err != nil {
			return nil, err
		}

		tp[i] = TagProbability{
			raw:         raw.(string),
			Probability: link.Probability,
		}
	}

	return tp, nil
}

// Step advances the model.
func (m *Model) Step() error {
	next, err := m.pickNext()
	if err != nil {
		return err
	}

	// Shift the past elements to the left to make room for the new word.
	size := len(m.past)
	copy(m.past, m.past[1:size])
	m.past[size-1] = next

	m.current = next

	return nil
}

func (m *Model) pickNext() (string, error) {
	links, err := m.nextLinks()
	if err != nil {
		return "", err
	}

	index := rand.Float64()
	var passed float64

	for _, link := range links {
		passed += link.Probability
		if passed > index {
			raw, err := m.chain.Get(link.ID)
			if err != nil {
				return "", err
			}

			return raw.(string), nil
		}
	}

	return "", errors.New("failed")
}

func (m *Model) nextLinks() ([]markov.Link, error) {
	id, err := m.chain.Find(strings.Join(m.past, " "))
	if err != nil {
		return nil, err
	}

	links, err := m.chain.Links(id)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		// If the chain ends in a unique phrase the chain will end.
		// Restart it at a random point. This isn't ideal, since it may
		// be mid-sentence.
		err = m.reseed()
		if err != nil {
			return nil, err
		}
		return m.nextLinks()
	}

	return links, nil
}

func (m *Model) reseed() error {
	seed, err := randomSeed(m.chain)
	if err != nil {
		return err
	}

	m.past = strings.Split(seed, " ")
	return nil
}

// TagProbability contains a Tag and the probability that it will be used next.
// Returned in a slice from Model.NextTags.
type TagProbability struct {
	raw         string
	Probability float64
}

// Tag parses the tag in it's raw form and returns it.
func (tp *TagProbability) Tag() Tag {
	return parseTag(tp.raw)
}
