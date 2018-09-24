package randtxt

import (
	"fmt"
	"strings"
)

type Tag struct {
	Text string
	POS  string
}

func (t Tag) IsZero() bool {
	return t.Text == "" && t.POS == ""
}

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
