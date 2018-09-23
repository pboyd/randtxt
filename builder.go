package randtxt

import (
	"fmt"
	"strings"

	"github.com/pboyd/markov"
)

type Builder struct {
	chain     markov.WriteChain
	ngramSize int
}

func NewBuilder(chain markov.WriteChain, ngramSize int) *Builder {
	return &Builder{
		chain:     chain,
		ngramSize: ngramSize,
	}
}

func (b *Builder) Feed(sources ...<-chan Tag) error {
	ngrams := make([]<-chan interface{}, len(sources))
	for i, source := range sources {
		ngrams[i] = b.joinTags(source)
	}

	return markov.Feed(b.chain, ngrams...)
}

func (b *Builder) joinTags(tags <-chan Tag) <-chan interface{} {
	ngrams := make(chan interface{})

	endOfSentence := true

	go func() {
		defer close(ngrams)

		ngram := make([]string, 0, b.ngramSize)

		for tag := range tags {
			switch tag.POS {
			case "-LRB-", "``", "-RRB-", "''", "SYM":
				continue
			}

			// Lower case words that begin sentences, unless it's a proper noun.
			if endOfSentence && (tag.Text != "I" && tag.POS != "NNP" && tag.POS != "NNPS") {
				tag.Text = strings.ToLower(tag.Text)
			}

			if tag.POS != "POS" && tag.POS != "VBZ" {
				tag.Text = strings.TrimLeft(tag.Text, "'")
			}

			endOfSentence = tag.POS == "."

			gram := fmt.Sprintf("%s/%s", tag.Text, tag.POS)

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
		}
	}()

	return ngrams
}
