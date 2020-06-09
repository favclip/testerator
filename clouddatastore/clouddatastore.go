package clouddatastore

import (
	"context"
	"os"
	"sync"

	clds "cloud.google.com/go/datastore"
	"github.com/favclip/testerator"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, func(s *testerator.Setup) error {
		return Cleanup(s.Context)
	})
}

func Cleanup(ctx context.Context) error {

	cdsCli, err := clds.NewClient(ctx, os.Getenv("DATASTORE_PROJECT_ID"))
	if err != nil {
		return err
	}

	q := clds.NewQuery("__namespace__").KeysOnly()
	namespaceKeys, err := cdsCli.GetAll(ctx, q, nil)
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

			q := clds.NewQuery("__kind__").Namespace(nsKey.Name).KeysOnly()
			kindKeys, err := cdsCli.GetAll(ctx, q, nil)
			if err != nil {
				rErr = err
				return
			}

			for _, kindKey := range kindKeys {
				q := clds.NewQuery(kindKey.Name).Namespace(kindKey.Namespace).KeysOnly()
				keys, err := cdsCli.GetAll(ctx, q, nil)
				if err != nil {
					rErr = err
					return
				}
				err = cdsCli.DeleteMulti(ctx, keys)
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
