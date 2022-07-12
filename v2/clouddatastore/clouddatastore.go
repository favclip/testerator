package clouddatastore

import (
	"context"
	"os"

	clds "cloud.google.com/go/datastore"
	"github.com/favclip/testerator/v2"
	"golang.org/x/sync/errgroup"
)

func init() {
	testerator.DefaultSetup.AppendCleanup(func(s *testerator.Setup) error {
		return Cleanup(s.Context)
	})
}

// Cleanup after test running.
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

	var eg errgroup.Group
	for _, nsKey := range namespaceKeys {
		nsKey := nsKey
		eg.Go(func() error {
			q := clds.NewQuery("__kind__").Namespace(nsKey.Name).KeysOnly()
			kindKeys, err := cdsCli.GetAll(ctx, q, nil)
			if err != nil {
				return err
			}

			for _, kindKey := range kindKeys {
				q := clds.NewQuery(kindKey.Name).Namespace(kindKey.Namespace).KeysOnly()
				keys, err := cdsCli.GetAll(ctx, q, nil)
				if err != nil {
					return err
				}

				const limit = 500
				for {
					if len(keys) <= limit {
						err = cdsCli.DeleteMulti(ctx, keys)
						if err != nil {
							return err
						}
						break
					}

					err = cdsCli.DeleteMulti(ctx, keys[0:limit])
					if err != nil {
						return err
					}

					keys = keys[limit:]
				}
			}

			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
