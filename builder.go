package randtxt

import (
	"strings"

	"github.com/pboyd/markov"
)

type ModelBuilder struct {
	chain     markov.WriteChain
	ngramSize int
	TagSet    TagSet
}

func NewModelBuilder(chain markov.WriteChain, ngramSize int) *ModelBuilder {
	return &ModelBuilder{
		chain:     chain,
		ngramSize: ngramSize,
		TagSet:    PennTreebankTagSet,
	}
}

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
