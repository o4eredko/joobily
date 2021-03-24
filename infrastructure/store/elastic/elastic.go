package elastic

import (
	"context"
	"fmt"
	stdLog "log"
	"os"

	"github.com/olivere/elastic/v7"
)

type (
	store struct {
		Client *elastic.Client
	}
)

func New(elasticURL string) *store {
	errorlog := stdLog.New(os.Stdout, "APP ", stdLog.LstdFlags)
	client, err := elastic.NewClient(
		elastic.SetURL(elasticURL),
		elastic.SetErrorLog(errorlog),
	)
	if err != nil {
		panic(err)
	}

	info, code, err := client.Ping(elasticURL).Do(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
	return &store{
		Client: client,
	}
}

func (s *store) Shutdown() {
	s.Client.Stop()
}
