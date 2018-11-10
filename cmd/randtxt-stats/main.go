package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/pboyd/markov"
	"github.com/pboyd/randtxt"
)

var (
	source      string
	entropyPath string
)

func init() {
	flag.StringVar(&source, "chain", "", "path to the chain file")
	flag.StringVar(&entropyPath, "entropy", "", "path to the entropy output file")
	flag.Parse()
}

func main() {
	if source == "" {
		fmt.Fprintf(os.Stderr, "error: -chain is required\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if entropyPath == "" {
		fmt.Fprintf(os.Stderr, "error: -entropy is required\n")
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

	outFh, err := os.Create(entropyPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to create entropy output file: %v\n", err)
		os.Exit(1)
	}

	defer outFh.Close()

	err = writeEntropy(outFh, chain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to find entropy: %v\n", err)
		os.Exit(2)
	}
}

// writeEntropy walks through the model and writes the entropy calculation for
// each step. writeEntropy stops when the mean of the generated values settles
// down.
func writeEntropy(w io.Writer, chain markov.Chain) error {
	model, err := randtxt.NewModel(chain, "")
	if err != nil {
		return err
	}

	sum := 0.0
	avg := 0.0

	for i := 0; ; i++ {
		err := model.Step()
		if err != nil {
			return err
		}

		p, err := model.NextTags()
		if err != nil {
			return err
		}

		e := singleEntropy(p)
		fmt.Fprintln(w, e)

		sum += e
		if i%1000 == 0 {
			lastAvg := avg
			avg = sum / float64(i)
			if math.Abs(lastAvg-avg) < 0.0001 {
				break
			}
		}
	}

	return nil
}

func singleEntropy(p []randtxt.TagProbability) float64 {
	sum := 0.0
	for _, tp := range p {
		sum += tp.Probability * (math.Log2(tp.Probability))
	}

	if sum == 0.0 {
		return 0.0
	}

	return -sum
}
