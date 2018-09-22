package randtxt

import "strings"

type Tag struct {
	Text string
	POS  string
}

func (t Tag) IsZero() bool {
	return t.Text == "" && t.POS == ""
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
