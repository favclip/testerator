package memcache

import (
	"github.com/favclip/testerator"
	"google.golang.org/appengine/memcache"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, cleanUpMemcache)
}

func cleanUpMemcache(s *testerator.Setup) error {
	memcache.Flush(s.Context)
	return nil
}
