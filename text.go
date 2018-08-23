package randtxt

import (
	"math/rand"
	"strings"
	"unicode"

	"github.com/pboyd/markov"
)

type Generator struct {
	Chain markov.Chain
}

// Paragraph returns a paragraph containing between "min" and "max" sentences.
func (g *Generator) Paragraph(min, max int) string {
	stop := make(chan struct{})
	defer close(stop)

	tags := genTags(stop, markov.RandomWalker(g.Chain, 0))

	sentences := make([]string, rand.Intn(max-min)+min)
	for i := range sentences {
		sentences[i] = genSentence(tags)
	}

	return strings.Join(sentences, " ")
}

func genSentence(tags <-chan tag) string {
	var sentence strings.Builder

	tag := <-tags
	tag.Text = titleCase(tag.Text)

	for {
		sentence.WriteString(tag.Text)

		if tag.Text == "." {
			break
		}

		tag = <-tags
		switch tag.Tag {
		case ".", ",", "POS":
		case "RB":
			if tag.Text == "n't" {
				continue
			}
			fallthrough
		default:
			sentence.WriteString(" ")
		}
	}

	return sentence.String()
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

func genTags(cancel <-chan struct{}, walker markov.Walker) <-chan tag {
	tags := make(chan tag)

	go func() {
		for {
			ngram, err := walker.Next()
			if err != nil {
				panic(err)
			}

			grams := strings.Split(ngram.(string), " ")
			for _, gram := range grams {
				split := strings.Split(gram, "/")
				tags <- tag{
					Text: split[0],
					Tag:  split[1],
				}
			}
		}
	}()

	return tags
}
