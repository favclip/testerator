package testerator

import (
	"testing"

	"google.golang.org/appengine/v2/datastore"
)

type Document struct {
	Name string
}

func TestPutDocument(t *testing.T) {
	_, c, err := SpinUp()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer SpinDown()

	_, c, err = SpinUp()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer SpinDown()

	doc := Document{Name: "sample"}
	key, err := datastore.Put(c, datastore.NewIncompleteKey(c, "document", nil), &doc)
	if err != nil {
		t.Fatal(err)
	}

	doc2 := Document{}
	if err = datastore.Get(c, key, &doc2); err != nil {
		t.Fatal(err)
	}

	if doc2.Name != "sample" {
		t.Fatalf("Name is not sample. got=%s", doc2.Name)
	}
}
