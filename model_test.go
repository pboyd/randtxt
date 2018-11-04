package randtxt

import (
	"reflect"
	"testing"
)

func TestModel(t *testing.T) {
	chain, close := testChain(t, "testfiles/ion/trigram.mkv")
	defer close()

	model, err := NewModel(chain, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var last []TagProbability
	for i := 0; i < 100; i++ {
		next, err := model.NextTags()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sum := 0.0
		for _, tp := range next {
			sum += tp.Probability
		}

		if int(sum+0.5) != 1 {
			t.Errorf("got %f, want %f", sum, 1.0)
		}

		if reflect.DeepEqual(last, next) {
			t.Errorf("got duplicate next tags")
		}

		last = next

		model.Step()
	}
}
