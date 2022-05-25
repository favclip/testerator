package datastore

import (
	"testing"

	"google.golang.org/appengine/v2"
	"google.golang.org/appengine/v2/aetest"
	"google.golang.org/appengine/v2/datastore"
)

func TestCleanup(t *testing.T) {
	opt := &aetest.Options{
		AppID:                       "cleanup-test",
		StronglyConsistentDatastore: true,
		SuppressDevAppServerLog:     true,
	}
	inst, err := aetest.NewInstance(opt)
	if err != nil {
		t.Fatal(err)
	}
	defer inst.Close()
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := appengine.NewContext(req)

	type Data struct {
		Str string
	}

	defaultNsKey := datastore.NewKey(ctx, "Test", "a", 0, nil)
	_, err = datastore.Put(ctx, defaultNsKey, &Data{Str: "default"})
	if err != nil {
		t.Fatal(err)
	}

	foobarCtx, err := appengine.Namespace(ctx, "foobar")
	if err != nil {
		t.Fatal(err)
	}
	foobarNsKey := datastore.NewKey(foobarCtx, "Test", "a", 0, nil)
	_, err = datastore.Put(foobarCtx, foobarNsKey, &Data{Str: "default"})
	if err != nil {
		t.Fatal(err)
	}

	err = Cleanup(ctx)
	if err != nil {
		t.Fatal(err)
	}

	err = datastore.Get(ctx, defaultNsKey, &Data{})
	if err == datastore.ErrNoSuchEntity {
		// ok
	} else {
		t.Fatal(err)
	}

	err = datastore.Get(foobarCtx, foobarNsKey, &Data{})
	if err == datastore.ErrNoSuchEntity {
		// ok
	} else {
		t.Fatal(err)
	}
}
