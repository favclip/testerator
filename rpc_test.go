package testerator

import (
	"testing"

	"google.golang.org/appengine/search"
)

type Sample struct {
	Foo string
}

func TestPutDocument(t *testing.T) {
	_, c, err := SpinUp()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer SpinDown()
	// exec twice
	_, c, err = SpinUp()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer SpinDown()

	idx, err := search.Open("test")
	if err != nil {
		t.Fatal(err.Error())
	}

	doc := &Sample{"Hi!"}
	_, err = idx.Put(c, "aaa", doc)
	if err != nil {
		t.Fatal(err.Error())
	}
}
