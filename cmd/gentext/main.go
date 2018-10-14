package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"

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
	flag.IntVar(&count, "count", 1, "number of paragraphs to generate")
	flag.IntVar(&seed, "seed", 0, "random seed")
	flag.Parse()
}

func main() {
	if seed == 0 {
		seed = os.Getpid()
		fmt.Fprintf(os.Stderr, "-seed=%d\n", seed)
	}
	rand.Seed(int64(seed))

	if source == "" {
		fmt.Fprintf(os.Stderr, "error: chain is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

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

	gen, err := randtxt.NewGenerator(chain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid chain: %v\n", err)
		os.Exit(2)
	}

	for i := 0; i < count; i++ {
		var err error
		err = gen.WriteParagraph(os.Stdout, 3, 6)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to generate paragraph: %v\n", err)
			os.Exit(2)
		}

		if i < count-1 {
			io.WriteString(os.Stdout, "\n\n")
		}
	}
	io.WriteString(os.Stdout, "\n")
}
