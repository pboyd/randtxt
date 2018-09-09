package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/pboyd/markov"
	"github.com/pboyd/randtxt"
)

var (
	source    string
	count     int
	seed      int
	ngramSize int
)

func init() {
	flag.StringVar(&source, "chain", "", "path to the chain file")
	flag.IntVar(&count, "count", 10, "number of paragraphs to generate")
	flag.IntVar(&seed, "seed", 0, "random seed")
	flag.IntVar(&ngramSize, "n", 3, "ngram size in chain")
	flag.Parse()
}

func main() {
	if seed == 0 {
		seed = os.Getpid()
		fmt.Fprintf(os.Stderr, "-seed=%d\n", seed)
	}
	rand.Seed(int64(seed))

	fh, err := os.Open(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", source, err)
		os.Exit(1)
	}
	defer fh.Close()

	chain, err := markov.ReadDiskChain(fh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read chain: %v\n", err)
		os.Exit(1)
	}

	gen := randtxt.NewGenerator(chain, ngramSize)

	paragraphs := make([]string, count)
	for i := range paragraphs {
		var err error
		paragraphs[i], err = gen.Paragraph(3, 6)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to generate paragraph: %v\n", err)
			os.Exit(2)
		}
	}
	fmt.Println(strings.Join(paragraphs, "\n\n"))
}
