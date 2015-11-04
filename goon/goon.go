package goon

import (
	"github.com/favclip/testerator"
	"github.com/mjibson/goon"
)

func init() {
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, goonCleanUp)
}

func goonCleanUp(s *testerator.Setup) error {
	goon.FromContext(s.Context).FlushLocalCache()
	return nil
}
