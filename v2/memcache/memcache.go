package memcache

import (
	"context"
	"fmt"
	"os"

	"github.com/favclip/testerator/v2"
	"google.golang.org/appengine/memcache"
)

func init() {
	testerator.DefaultSetup.AppendCleanup(func(s *testerator.Setup) error {
		if s.Disable1stGen {
			_, _ = fmt.Fprintln(os.Stderr, `don't use "github.com/favclip/testerator/v2/memcache" package with Disable1stGen`)
		}
		return Cleanup(s.Context)
	})
}

func Cleanup(ctx context.Context) error {
	return memcache.Flush(ctx)
}
