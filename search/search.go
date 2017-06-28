package search

import (
	"context"
	"errors"

	"github.com/favclip/testerator"
	searchpb "github.com/favclip/testerator/aeinternal/search"
	"github.com/golang/protobuf/proto"
	netcontext "golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/search"
)

var ctxKey = "https://code.google.com/p/googleappengine/issues/detail?id=12747"

var ErrSetupRequired = errors.New("please use '_ \"github.com/favclip/testerator/search\"'")

type searchSniffer struct {
	searchIndexDocumentRequests []*searchpb.IndexDocumentRequest
}

func init() {
	testerator.DefaultSetup.Setuppers = append(testerator.DefaultSetup.Setuppers, setup)
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, cleanup)
}

func setup(s *testerator.Setup) error {
	if sniff, ok := s.Context.Value(&ctxKey).(*searchSniffer); ok {
		sniff.searchIndexDocumentRequests = nil
		return nil
	}

	sniff := &searchSniffer{}
	s.Context = context.WithValue(s.Context, &ctxKey, sniff)
	s.Context = appengine.WithAPICallFunc(s.Context, func(ctx netcontext.Context, service, method string, in, out proto.Message) error {
		if service == "search" && method == "IndexDocument" {
			b, err := proto.Marshal(in)
			if err != nil {
				return err
			}

			req := &searchpb.IndexDocumentRequest{}
			err = proto.Unmarshal(b, req)
			if err != nil {
				return err
			}

			sniff.searchIndexDocumentRequests = append(sniff.searchIndexDocumentRequests, req)
		}
		return appengine.APICall(ctx, service, method, in, out)
	})

	return nil
}

func cleanup(s *testerator.Setup) error {
	c := s.Context
	sniff, ok := c.Value(&ctxKey).(*searchSniffer)
	if !ok {
		return ErrSetupRequired
	}

	indexNames := make(map[string]bool, 0)
	for _, req := range sniff.searchIndexDocumentRequests {
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

	sniff.searchIndexDocumentRequests = nil

	return nil
}
