package testerator

import (
	"os"
	"runtime"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
)

type Helper func(s *Setup) error

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

var DefaultSetup *Setup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	DefaultSetup = &Setup{}
}

func SpinUp() (aetest.Instance, context.Context, error) {
	err := DefaultSetup.SpinUp()
	return DefaultSetup.Instance, DefaultSetup.Context, err
}

func SpinDown() error {
	return DefaultSetup.SpinDown()
}

func IsCI() bool {
	return os.Getenv("CI") != ""
}

func (s *Setup) SpinUp() error {
	s.Lock()
	defer s.Unlock()

	if s.ResetThreshold == 0 {
		s.ResetThreshold = 15
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
