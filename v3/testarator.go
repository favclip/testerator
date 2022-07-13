package testerator

import (
	"context"
	"os"
	"sync"

	"google.golang.org/appengine/v2/aetest"
)

// Helper uses for setup hooks to Setup struct.
type Helper func(s *Setup) error

// Setup contains aetest.Instance and other environment for setup and clean up.
type Setup struct {
	Instance       aetest.Instance
	Disable1stGen  bool
	RaisePanic     bool
	Options        *aetest.Options
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
	// runtime.GOMAXPROCS(runtime.NumCPU())
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

	opt := s.Options
	if opt == nil {
		opt = &aetest.Options{
			AppID:                       "unittest",
			StronglyConsistentDatastore: true,
			SuppressDevAppServerLog:     true,
		}
	}

	if s.Disable1stGen {
		s.Instance = &mockAEInstance{}
		s.Context = context.Background()
	} else {
		inst, err := aetest.NewInstance(opt)
		if err != nil {
			return err
		}
		req, err := inst.NewRequest("GET", "/", nil)
		if err != nil {
			return err
		}

		c := appengine.NewContext(req)

		s.Instance = inst
		s.Context = c
	}

	for _, setupper := range s.Setuppers {
		err := setupper(s)
		if err != nil {
			return err
		}
	}

	return nil
}

// SpinDown dev server.
//
// This function clean up dev server environment.
// call each DefaultSetup.Cleaners. usually, it means cleanup Datastore and Search APIs and miscs.
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

	// clean up environment
	for _, c := range s.Cleaners {
		err := c(s)
		if err != nil {
			if s.RaisePanic {
				panic(err)
			}
			return err
		}
	}

	s.counter--

	return nil
}

func (s *Setup) AppendSetuppers(h Helper) {
	s.Lock()
	defer s.Unlock()

	s.Setuppers = append(s.Setuppers, h)
}

func (s *Setup) AppendCleanup(h Helper) {
	s.Lock()
	defer s.Unlock()

	s.Cleaners = append(s.Cleaners, h)
}
