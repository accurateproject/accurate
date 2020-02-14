package main

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/accurateproject/accurate/utils"
	elastic "gopkg.in/olivere/elastic.v5"
)

// hopa ceva descriere
type ElasticService struct {
	ctx           context.Context
	client        *elastic.Client
	bulkProcessor *elastic.BulkProcessor
	indexCache    utils.StringMap
}

func NewElasticService(urls []string, user, pass string) (*ElasticService, error) {
	ctx := context.Background()
	var options elastic.ClientOptionFunc
	if len(urls) != 0 {
		options = append(options, elastic.SetURL(urls...))
	}
	if user != "" || pass != "" {
		options = append(options, elastic.SetBasicAuth(user, pass))
	}
	client, err := elastic.NewClient(options...)
	if err != nil {
		return err
	}

	// Setup a bulk processor
	bp, err := client.BulkProcessor().
		Name("cc-elastic-cdr-bulk-processor").
		Workers(8).
		BulkActions(1000).               // commit if # requests >= 1000
		BulkSize(2 << 20).               // commit if size of requests >= 2 MB
		FlushInterval(30 * time.Second). // commit every 30s
		Do(context.Background())
	if err != nil {
		return err
	}
	return &ElasticService{
		ctx:           ctx,
		client:        client,
		bulkProcessor: bp,
		indexCache:    make(utils.StringMap),
	}
}

func (es *ElasticService) createCDRIndex(indexName string) error {
	// Use the IndexExists service to check if a specified index exists.
	exists, err := es.client.IndexExists(indexName).Do(es.ctx)
	if err != nil {
		utils.Logger.Error("error checking elastic index.", zap.Error(err))
		return err
	}
	if !exists {
		// Create a new index.
		createIndex, err := es.client.CreateIndex(indexName).Body(`{
                 "mappings" : {
                            "key" : {
                                "properties" : {
                                    "key" : { "type" : "keyword", "index": true },
                                    "value" : { "type" : "keyword", "index": true },
                                    "destination": { "type": "keyword", "index": true },
                                    "usage" : { "type": "long" },
                                    "cost"  : { "type": "float" },
                                    "time" : { "type" : "date" }
                                }
                            }
                        }
                    }`).Do(es.ctx)
		if err != nil {
			utils.Logger.Error("error creating elastic index.", zap.Error(err))
			return err
		}
		if !createIndex.Acknowledged {
			utils.Logger.Warn("createIndex not acknowledged!")
		}
		utils.Logger.Debug("createIndex: ", zap.Any("create index response", createIndex))
	}
	return nil
}

func (es *ElasticService) IndexCDR(cdr *CDR) {
	r := elastic.NewBulkIndexRequest().
		Index("keys").
		Type("key").
		Id(index).
		Doc(cdr)

	es.bulkProcessor.Add(r)
}

func hopa(
/*for i := 0; i < CYCLES; i++ {

	index := strconv.Itoa(i)
	termQuery := elastic.NewTermQuery("key", "key"+index)
	get, err := client.Search().
		Index("keys").    // search in index "twitter"
		Query(termQuery). // specify the query
		Do(ctx)           // execute
	if err != nil {
		log.Fatal(err)
	}
	if i%10000 == 0 {
		log.Printf("get: %+v", get)
	}
}*/
)
