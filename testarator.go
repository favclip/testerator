package testerator

import (
	"os"
	"runtime"
	"sync"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
	searchpb "google.golang.org/appengine/internal/search"
	"google.golang.org/appengine/memcache"
	"google.golang.org/appengine/search"
)

type Cleanup func(s *Setup) error

type Setup struct {
	Instance        aetest.Instance
	Context         context.Context
	originalContext context.Context
	counter         int
	Cleaners        []Cleanup
	total           int
	ResetThreshold  int
	SpinDowns       []chan struct{}

	searchIndexDocumentRequests []*searchpb.IndexDocumentRequest

	sync.Mutex
}

var DefaultSetup *Setup

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	DefaultSetup = &Setup{}
	DefaultSetup.Cleaners = append(DefaultSetup.Cleaners, cleanUpDatastore)
	DefaultSetup.Cleaners = append(DefaultSetup.Cleaners, cleanUpMemcache)
	DefaultSetup.Cleaners = append(DefaultSetup.Cleaners, cleanUpSearchDocument)
}

func SpinUp() (aetest.Instance, context.Context, error) {
	err := DefaultSetup.SpinUp()
	return DefaultSetup.Instance, DefaultSetup.Context, err
}

func SpinDown() error {
	return DefaultSetup.SpinDown()
}

func IsCI() bool {
	return os.Getenv("CIRCLECI") != ""
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
		s.searchIndexDocumentRequests = nil // reset
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
	s.originalContext = c
	s.Context = appengine.WithAPICallFunc(c, func(ctx context.Context, service, method string, in, out proto.Message) error {
		if service == "search" && method == "IndexDocument" {
			docReq := in.(*searchpb.IndexDocumentRequest)
			s.searchIndexDocumentRequests = append(s.searchIndexDocumentRequests, docReq)
		}
		return appengine.APICall(c, service, method, in, out)
	})

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

		closeInstane := func() {
			ch := make(chan struct{})
			s.SpinDowns = append(s.SpinDowns, ch)
			go func(inst aetest.Instance) {
				defer func() { ch <- struct{}{} }()
				inst.Close()
			}(s.Instance)
		}

		if s.counter == 0 {
			closeInstane()
			s.Instance = nil
		} else if s.total%s.ResetThreshold == 0 {
			// Sometimes spin down causes. avoid to saturate file descriptor.
			closeInstane()
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

func cleanUpDatastore(s *Setup) error {
	t := datastore.NewQuery("__kind__").KeysOnly().Run(s.originalContext)
	kinds := make([]string, 0)
	for {
		key, err := t.Next(nil)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return err
		}
		kinds = append(kinds, key.StringID())
	}

	for _, kind := range kinds {
		q := datastore.NewQuery(kind).KeysOnly()
		keys, err := q.GetAll(s.originalContext, nil)
		if err != nil {
			return err
		}
		err = datastore.DeleteMulti(s.originalContext, keys)
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanUpMemcache(s *Setup) error {
	memcache.Flush(s.originalContext)
	return nil
}

func cleanUpSearchDocument(s *Setup) error {
	c := s.originalContext
	indexNames := make(map[string]bool, 0)
	for _, req := range s.searchIndexDocumentRequests {
		indexNames[*req.GetParams().GetIndexSpec().Name] = true
	}
	for indexName, _ := range indexNames {
		idx, err := search.Open(indexName)
		if err != nil {
			return err
		}
		iter := idx.List(c, &search.ListOptions{IDsOnly: true})
		for {
			docID, err := iter.Next(nil)
			if err == search.Done {
				break
			} else if err != nil {
				return err
			}
			err = idx.Delete(c, docID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
