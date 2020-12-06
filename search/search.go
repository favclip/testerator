package search

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/favclip/testerator"
	searchpb "github.com/favclip/testerator/aeinternal/search"
	"github.com/golang/protobuf/proto"
	netcontext "golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/search"
)

type ctxKey struct{}

var ErrSetupRequired = errors.New("please use '_ \"github.com/favclip/testerator/search\"'")

func init() {
	testerator.DefaultSetup.Setuppers = append(testerator.DefaultSetup.Setuppers, func(s *testerator.Setup) error {
		if s.Disable1stGen {
			_, _ = fmt.Fprintln(os.Stderr, `don't use "github.com/favclip/testerator/search" package with Disable1stGen`)
		}
		return Setup(s.Context)
	})
	testerator.DefaultSetup.Cleaners = append(testerator.DefaultSetup.Cleaners, func(s *testerator.Setup) error {
		if s.Disable1stGen {
			_, _ = fmt.Fprintln(os.Stderr, `don't use "github.com/favclip/testerator/search" package with Disable1stGen`)
		}
		return Cleanup(s.Context)
	})
}

type SearchSniffer struct {
	IndexDocumentRequests []*searchpb.IndexDocumentRequest
}

func Setup(ctx context.Context) error {
	if sniff, ok := ctx.Value(ctxKey{}).(*SearchSniffer); ok {
		sniff.IndexDocumentRequests = nil
		return nil
	}

	sniff := &SearchSniffer{}
	ctx = context.WithValue(ctx, ctxKey{}, sniff)
	ctx = appengine.WithAPICallFunc(ctx, func(ctx netcontext.Context, service, method string, in, out proto.Message) error {
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

			sniff.IndexDocumentRequests = append(sniff.IndexDocumentRequests, req)
		}
		return appengine.APICall(ctx, service, method, in, out)
	})

	return nil
}

func Cleanup(ctx context.Context) error {
	sniff, ok := ctx.Value(ctxKey{}).(*SearchSniffer)
	if !ok {
		return ErrSetupRequired
	}

	indexNames := make(map[string]bool, 0)
	for _, req := range sniff.IndexDocumentRequests {
		indexNames[*req.GetParams().GetIndexSpec().Name] = true
	}
	for indexName, _ := range indexNames {
		idx, err := search.Open(indexName)
		if err != nil {
			return err
		}
		iter := idx.List(ctx, &search.ListOptions{IDsOnly: true})
		for {
			docID, err := iter.Next(nil)
			if err == search.Done {
				break
			} else if err != nil {
				return err
			}
			err = idx.Delete(ctx, docID)
			if err != nil {
				return err
			}
		}
	}

	sniff.IndexDocumentRequests = nil

	return nil
}
