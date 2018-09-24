package randtxt

import "testing"

func TestPennTreebankJoin(t *testing.T) {
	cases := []struct {
		tag, prev string
		expected  string
	}{
		{
			tag:      "Paul/NNP",
			prev:     "John/NNP",
			expected: " Paul",
		},
		{
			tag:      "Paul/NNP",
			prev:     "",
			expected: "Paul",
		},
		{
			tag:      "./.",
			prev:     "ran/VB",
			expected: ".",
		},
		{
			tag:      "he/PP",
			prev:     "./.",
			expected: " He",
		},
		{
			tag:      "'s/POS",
			prev:     "he/PP",
			expected: "'s",
		},
	}

	for i, c := range cases {
		tag := parseTag(c.tag)
		prev := parseTag(c.prev)
		actual := PennTreebankTagSet.Join(tag, prev)
		if actual != c.expected {
			t.Errorf("%d: got %q, want %q", i, actual, c.expected)
		}
	}
}

func TestPennTreebankNormalize(t *testing.T) {
	cases := []struct {
		tag, prev string
		expected  string
	}{
		{
			tag:      "Was/VBD",
			prev:     "",
			expected: "was/VBD",
		},
		{
			tag:      "Was/VBD",
			prev:     "./.",
			expected: "was/VBD",
		},
		{
			tag:      "Ringo/NNP",
			prev:     "./.",
			expected: "Ringo/NNP",
		},
		{
			tag:      "(/-LRB-",
			prev:     "",
			expected: "",
		},
		{
			tag:      "'of/IN",
			prev:     "",
			expected: "of/IN",
		},
		{
			tag:      "'s/POS",
			prev:     "she/PP",
			expected: "'s/POS",
		},
	}

	for i, c := range cases {
		tag := parseTag(c.tag)
		prev := parseTag(c.prev)

		actual := PennTreebankTagSet.Normalize(tag, prev).String()

		if actual != c.expected {
			t.Errorf("%d: got %q, want %q", i, actual, c.expected)
		}
	}
}
