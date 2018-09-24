package randtxt

import (
	"strings"

	"github.com/pboyd/markov"
)

// ModelBuilder builds a model that Generator can use.
type ModelBuilder struct {
	chain     markov.WriteChain
	ngramSize int
	TagSet    TagSet
}

// NewModelBuilder creates a ModelBuilder instance.
//
// The model will be written to "chain".
//
// ngramSize is the number of words to include in each ngram. Must be greater
// than 1.
//
// See cmd/readtsv for an example.
func NewModelBuilder(chain markov.WriteChain, ngramSize int) *ModelBuilder {
	return &ModelBuilder{
		chain:     chain,
		ngramSize: ngramSize,
		TagSet:    PennTreebankTagSet,
	}
}

// Feed reads tags from one or more channels and writes them to the output
// chain.
func (b *ModelBuilder) Feed(sources ...<-chan Tag) error {
	ngrams := make([]<-chan interface{}, len(sources))
	for i, source := range sources {
		ngrams[i] = b.joinTags(source)
	}

	return markov.Feed(b.chain, ngrams...)
}

func (b *ModelBuilder) joinTags(tags <-chan Tag) <-chan interface{} {
	ngrams := make(chan interface{})

	go func() {
		defer close(ngrams)

		var prev Tag
		ngram := make([]string, 0, b.ngramSize)

		for tag := range tags {
			tag = b.TagSet.Normalize(tag, prev)
			if tag.Text == "" {
				continue
			}

			gram := tag.String()

			if len(ngram) < b.ngramSize {
				ngram = append(ngram, gram)

				if len(ngram) < b.ngramSize {
					continue
				}
			} else {
				ngrams <- gram

				copy(ngram[0:], ngram[1:])
				ngram[b.ngramSize-1] = gram
			}

			ngrams <- strings.Join(ngram, " ")

			prev = tag
		}
	}()

	return ngrams
}
