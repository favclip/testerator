package datastore

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/favclip/testerator"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func init() {
	testerator.DefaultSetup.AppendCleanup(func(s *testerator.Setup) error {
		if s.Disable1stGen {
			_, _ = fmt.Fprintln(os.Stderr, `don't use "github.com/favclip/testerator/datastore" package with Disable1stGen`)
		}
		return Cleanup(s.Context)
	})
}

func Cleanup(ctx context.Context) error {

	q := datastore.NewQuery("__namespace__").KeysOnly()
	namespaceKeys, err := q.GetAll(ctx, nil)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(namespaceKeys))

	var rErr error
	for _, nsKey := range namespaceKeys {

		nsKey := nsKey

		go func() {
			defer wg.Done()

			ctx, err := appengine.Namespace(ctx, nsKey.StringID())
			if err != nil {
				rErr = err
				return
			}

			q := datastore.NewQuery("__kind__").KeysOnly()
			kindKeys, err := q.GetAll(ctx, nil)

			for _, kindKey := range kindKeys {
				q := datastore.NewQuery(kindKey.StringID()).KeysOnly()
				keys, err := q.GetAll(ctx, nil)
				if err != nil {
					rErr = err
					return
				}
				err = datastore.DeleteMulti(ctx, keys)
				if err != nil {
					rErr = err
					return
				}
			}
		}()
	}

	wg.Wait()

	if rErr != nil {
		return rErr
	}

	return nil
}
