package datastore

import (
	"context"

	"github.com/favclip/testerator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, cleanup)
}

func cleanup(s *testerator.Setup) error {
	contexts := []context.Context{s.Context}

	q := datastore.NewQuery("__namespace__").KeysOnly()
	namespaces, err := q.GetAll(s.Context, nil)
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		nsContext, err := appengine.Namespace(s.Context, ns.StringID())
		if err != nil {
			return err
		}
		contexts = append(contexts, nsContext)
	}

	for _, ctx := range contexts {
		err := cleanupUnderContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanupUnderContext(ctx context.Context) error {
	t := datastore.NewQuery("__kind__").KeysOnly().Run(ctx)
	kinds := make([]string, 0)
	for {
		key, err := t.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		kinds = append(kinds, key.StringID())
	}

	for _, kind := range kinds {
		q := datastore.NewQuery(kind).KeysOnly()
		keys, err := q.GetAll(ctx, nil)
		if err != nil {
			return err
		}
		err = datastore.DeleteMulti(ctx, keys)
		if err != nil {
			return err
		}
	}

	return nil
}
