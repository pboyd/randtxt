package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pboyd/markov"
	"github.com/pboyd/randtxt"
)

var (
	output string
	update bool
	onDisk bool
	n      int
)

func init() {
	flag.StringVar(&output, "chain", "", "path the the output chain file")
	flag.BoolVar(&update, "update", false, "update the output file instead of overwriting it")
	flag.BoolVar(&onDisk, "disk", false, "write the chain directly to disk")
	flag.IntVar(&n, "n", 3, "ngram size")
	flag.Parse()
}

func main() {
	if output == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sources := flag.Args()
	if len(sources) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] source [source]...\n", os.Args[0])
		os.Exit(1)
	}

	tags := make([]<-chan randtxt.Tag, len(sources))

	for i, source := range sources {
		var err error
		tags[i], err = readTSV(source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "file error (%s): %v\n", source, err)
			os.Exit(1)
		}
	}

	diskChain, err := openOutputFile(output, update)
	if err != nil {
		fmt.Fprintf(os.Stderr, "file error (%s): %v\n", output, err)
		os.Exit(1)
	}

	if onDisk {
		builder := randtxt.NewModelBuilder(diskChain, n)
		err := builder.Feed(tags...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}
	} else {
		memoryChain := &markov.MemoryChain{}
		builder := randtxt.NewModelBuilder(memoryChain, n)
		err := builder.Feed(tags...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}

		err = markov.Copy(diskChain, memoryChain)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error copying chain to disk: %v\n", err)
			os.Exit(2)
		}
	}
}

func openOutputFile(path string, update bool) (markov.WriteChain, error) {
	if update {
		exists, err := fileExists(path)
		if err != nil {
			return nil, err
		}

		if !exists {
			update = false
		}
	}

	fh, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	if update {
		return markov.OpenDiskChainWriter(fh)
	}

	return markov.NewDiskChainWriter(fh)
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	// Some other error, probably an invalid path.
	return false, err
}

func readTSV(path string) (<-chan randtxt.Tag, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fh)

	tags := make(chan randtxt.Tag)

	go func() {
		defer func() {
			fh.Close()
			close(tags)
		}()

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "error reading file: %v", err)
				}
				break
			}

			tag := parseLine(line)
			if tag.Text == "" || tag.POS == "" {
				continue
			}

			tags <- tag
		}
	}()

	return tags, nil
}

func parseLine(line string) randtxt.Tag {
	line = strings.TrimSpace(line)

	i := strings.IndexByte(line, '\t')
	if i < 0 {
		return randtxt.Tag{}
	}

	return randtxt.Tag{
		Text: line[:i],
		POS:  line[i+1:],
	}
}
