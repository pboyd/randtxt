# randtxt [![GoDoc](https://godoc.org/github.com/golang/gddo?status.svg)](https://godoc.org/github.com/pboyd/randtxt)

Generates random text from Markov chains of tagged source text.

An example chain is included which was derived from [Plato's
Ion](https://www.gutenberg.org/ebooks/1635):

```sh
$ go get github.com/pboyd/randtxt
$ go run github.com/pboyd/randtxt/cmd/gentext -chain $GOPATH/src/github.com/pboyd/randtxt/testfiles/ion/trigram.mkv
```

> Have you already forgotten what you were saying? A rhapsode ought to
> interpret the mind of the poet. For the rhapsode ought to interpret the mind
> of the poet. For the poet is a light and winged and holy thing, and there is
> Phanosthenes of Andros, and Heraclides of Clazomenae, whom they have also
> appointed to the command of their armies and to other offices, although
> aliens, after they had shown their merit. And will they not choose Ion the
> Ephesian to be their general, and honour him, if he prove himself worthy?

To build a chain, use the [Stanford POS
Tagger](https://nlp.stanford.edu/software/tagger.shtml) to generate tagged
text, then run `cmd/readtsv`. For example:

```sh
go run github.com/pboyd/randtxt/cmd/readtsv -chain output.mkv $GOPATH/src/github.com/pboyd/randtxt/testfiles/ion/tagged.tsv
```

# License

This package is released under the terms of the Apache 2.0 license. See LICENSE.TXT.
