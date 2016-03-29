package datastore

import (
	"github.com/favclip/testerator"
	"google.golang.org/appengine/datastore"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, cleanup)
}

func cleanup(s *testerator.Setup) error {
	t := datastore.NewQuery("__kind__").KeysOnly().Run(s.Context)
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
		keys, err := q.GetAll(s.Context, nil)
		if err != nil {
			return err
		}
		err = datastore.DeleteMulti(s.Context, keys)
		if err != nil {
			return err
		}
	}

	return nil
}
