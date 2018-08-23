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
	source string
	count  int
	seed   int
)

func init() {
	flag.StringVar(&source, "chain", "", "path to the chain file")
	flag.IntVar(&count, "count", 10, "number of paragraphs to generate")
	flag.IntVar(&seed, "seed", 0, "random seed")
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

	chain, err := markov.ReadDiskChain(fh)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to read chain: %v\n", err)
		os.Exit(1)
	}

	gen := &randtxt.Generator{Chain: chain}

	paragraphs := make([]string, count)
	for i := range paragraphs {
		paragraphs[i] = gen.Paragraph(3, 6)
	}
	fmt.Println(strings.Join(paragraphs, "\n\n"))
}
