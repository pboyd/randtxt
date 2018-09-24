package randtxt

import (
	"bytes"
	"strings"
	"unicode"
)

type TagSet interface {
	// Join returns the text from "tag" prepended with the separator that
	// should be between "prev" and "tag".
	//
	// "prev" is the zero tag at the beginning of the text.
	Join(tag, prev Tag) string

	// Normalize converts "tag" to a consistent form. If the returned tag
	// text is blank the tag is ignored.
	Normalize(tag, prev Tag) Tag
}

var PennTreebankTagSet = pennTreebankTagSet{}

type pennTreebankTagSet struct{}

func (pt pennTreebankTagSet) Join(this, prev Tag) string {
	buf := bytes.NewBuffer(make([]byte, 0, len(this.Text)+1))

	needSpace := false

	switch this.POS {
	case ".", ",", ":", "POS":
	case "RB":
		if this.Text != "n't" {
			needSpace = true
		}
	case "VBZ":
		if !strings.HasPrefix(this.Text, "'") {
			needSpace = true
		}
	default:
		needSpace = true
	}

	if !prev.IsZero() && needSpace {
		buf.WriteString(" ")
	}

	word := this.Text

	switch prev.POS {
	case "", ".", ":":
		if prev.Text != ";" {
			word = titleCase(word)
		}
	}

	buf.WriteString(word)
	return buf.String()
}

func titleCase(text string) string {
	buf := []rune(text)
	if len(buf) == 0 {
		return ""
	}

	buf[0] = unicode.ToTitle(buf[0])
	return string(buf)
}

func (pt pennTreebankTagSet) Normalize(tag, prev Tag) Tag {
	switch tag.POS {
	case "-LRB-", "``", "-RRB-", "''", "SYM":
		return Tag{}
	}

	// Lower case words that begin sentences, unless it's a proper noun.
	if prev.IsZero() || prev.POS == "." {
		if tag.Text != "I" && tag.POS != "NNP" && tag.POS != "NNPS" {
			tag.Text = strings.ToLower(tag.Text)
		}
	}

	if tag.POS != "POS" && tag.POS != "VBZ" {
		tag.Text = strings.TrimLeft(tag.Text, "'")
	}

	return tag
}
