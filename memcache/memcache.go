package memcache

import (
	"context"

	"github.com/favclip/testerator"
	"google.golang.org/appengine/memcache"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, func(s *testerator.Setup) error {
		return Cleanup(s.Context)
	})
}

func Cleanup(ctx context.Context) error {
	return memcache.Flush(ctx)
}
