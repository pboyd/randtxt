package randtxt

import (
	"fmt"
	"strings"
)

// Tag represents a single tagged word.
type Tag struct {
	Text string

	// POS is the part of speech tag for the text.
	POS string
}

// IsZero tests if the tag is the empty zero value.
func (t Tag) IsZero() bool {
	return t.Text == "" && t.POS == ""
}

// String returns the tag in "Text/POS" form.
func (t Tag) String() string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s/%s", t.Text, t.POS)
}

func parseTag(gram string) Tag {
	split := strings.Split(gram, "/")
	if len(split) < 2 {
		return Tag{}
	}

	return Tag{
		Text: split[0],
		POS:  split[1],
	}
}
