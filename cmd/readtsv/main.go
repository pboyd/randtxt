package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pboyd/markov"
)

var (
	output string
	update bool
	onDisk bool
	n      int
)

const (
	phraseStart = "./."
)

func init() {
	flag.StringVar(&output, "chain", "", "path the the output chain file")
	flag.BoolVar(&update, "update", false, "update the output file instead of overwriting it")
	flag.BoolVar(&onDisk, "disk", false, "write the chain directly to disk")
	flag.IntVar(&n, "n", 2, "ngram size")
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

	ngrams := make([]<-chan interface{}, len(sources))

	for i, source := range sources {
		var err error
		ngrams[i], err = readTSV(source, n)
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
		err := markov.Feed(diskChain, ngrams...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error building chain: %v\n", err)
			os.Exit(2)
		}
	} else {
		memoryChain := &markov.MemoryChain{}
		err := markov.Feed(memoryChain, ngrams...)
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

func readTSV(path string, n int) (<-chan interface{}, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(fh)

	ngrams := make(chan interface{})

	go func() {
		defer func() {
			fh.Close()
			close(ngrams)
		}()

		current := make([]string, 0, n)
		emit := func() {
			if len(current) == 0 {
				return
			}

			ngrams <- fmt.Sprintf(strings.Join(current, " "))
			current = current[:0]
		}

		lastWasPhraseStart := false

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "error reading file: %v", err)
				}
				break
			}

			text, tag := parseLine(line)
			if text == "" {
				continue
			}

			switch tag {
			case ":", ".", "``", "-LRB-":
				if lastWasPhraseStart {
					continue
				}

				emit()
				ngrams <- phraseStart
				lastWasPhraseStart = true
				continue
			case "-RRB-", "''", "SYM":
				continue
			default:
				lastWasPhraseStart = false
			}

			if text != "I" && tag != "NNP" && tag != "NNPS" {
				text = strings.ToLower(text)
			}

			if tag != "POS" && tag != "VBZ" {
				text = strings.TrimLeft(text, "'")
			}

			current = append(current, fmt.Sprintf("%s/%s", text, tag))

			if len(current) == n {
				emit()
			}
		}
	}()

	return ngrams, nil
}

func parseLine(line string) (word string, tag string) {
	line = strings.TrimSpace(line)

	i := strings.IndexByte(line, '\t')
	if i < 0 {
		return
	}
	word = line[:i]
	tag = line[i+1:]

	return
}
