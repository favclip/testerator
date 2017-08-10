package testerator

import (
	"context"
	"os"
	"runtime"
	"sync"

	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

// Helper uses for setup hooks to Setup struct.
type Helper func(s *Setup) error

// Setup contains aetest.Instance and other environment for setup and clean up.
type Setup struct {
	Instance       aetest.Instance
	Context        context.Context
	counter        int
	Setuppers      []Helper
	Cleaners       []Helper
	total          int
	ResetThreshold int
	SpinDowns      []chan struct{}

	sync.Mutex
}

// DefaultSetup uses from bare functions.
var DefaultSetup *Setup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	DefaultSetup = &Setup{}
}

// SpinUp dispatch spin up request to DefaultSetup.SpinUp.
func SpinUp() (aetest.Instance, context.Context, error) {
	err := DefaultSetup.SpinUp()
	return DefaultSetup.Instance, DefaultSetup.Context, err
}

// SpinDown dispatch spin down request to DefaultSetup.SpinDown.
func SpinDown() error {
	return DefaultSetup.SpinDown()
}

// IsCI returns this execution environment is Continuous Integration server or not.
// Deprecated.
func IsCI() bool {
	return os.Getenv("CI") != ""
}

// SpinUp dev server.
//
// If you call this function twice. launch dev server and increment internal counter twice.
// 1st time then dev server is up and increment internal counter.
// 2nd time then dev server is increment internal counter only.
// see document for SpinDown function.
func (s *Setup) SpinUp() error {
	s.Lock()
	defer s.Unlock()

	if s.ResetThreshold == 0 {
		// NOTE
		//   https://cloud.google.com/appengine/docs/standard/go/release-notes
		//   August 9, 2017 Updated Go SDK to version 1.9.57.
		//   The aetest package now reuses HTTP connections, fixing a bug that exhausted file descriptors when running tests.
		s.ResetThreshold = 1000
	}

	s.total++
	s.counter++

	if s.Instance != nil {
		return nil
	}

	opt := &aetest.Options{AppID: "unittest", StronglyConsistentDatastore: true}
	inst, err := aetest.NewInstance(opt)
	if err != nil {
		return err
	}
	req, err := inst.NewRequest("GET", "/", nil)
	if err != nil {
		return err
	}

	c := appengine.NewContext(req)

	if err != nil {
		return err
	}

	s.Instance = inst
	s.Context = c

	for _, setupper := range s.Setuppers {
		err = setupper(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// SpinDown dev server.
//
// This function clean up dev server environment.
// However, internally there are two types of processing.
// #1. if internal counter == 0, spin down dev server simply.
// #2. otherwise, call each DefaultSetup.Cleaners. usually, it means cleanup Datastore and Search APIs.
// see document for SpinUp function.
func (s *Setup) SpinDown() error {
	s.Lock()
	defer s.Unlock()
	defer func() {
		if s.counter == 0 {
			for _, sd := range s.SpinDowns {
				<-sd
			}
			s.SpinDowns = nil
		}
	}()
	defer func() {
		if s.Instance == nil {
			return
		}

		closeInstance := func() {
			ch := make(chan struct{})
			s.SpinDowns = append(s.SpinDowns, ch)
			go func(inst aetest.Instance) {
				defer func() {
					ch <- struct{}{}
				}()
				inst.Close()
			}(s.Instance)
		}

		if s.counter == 0 {
			closeInstance()
			s.Instance = nil
		} else if s.total%s.ResetThreshold == 0 {
			// Sometimes spin down causes. avoid to saturate file descriptor.
			closeInstance()
			s.Instance = nil
		}
	}()

	s.counter--

	if s.counter == 0 {
		// server spin downed.
		return nil
	}

	// clean up environment
	for _, c := range s.Cleaners {
		err := c(s)
		if err != nil {
			return err
		}
	}

	return nil
}
