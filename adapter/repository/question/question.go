package question

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/olivere/elastic/v7"

	"gitlab.jooble.com/marketing_tech/joobily/domain"
	"gitlab.jooble.com/marketing_tech/joobily/domain/dto"
	"gitlab.jooble.com/marketing_tech/joobily/domain/entity"
	"gitlab.jooble.com/marketing_tech/joobily/usecase/question"
)

type (
	repo struct {
		indexName string
		client    *elastic.Client
		questions map[string]string
		mu        sync.RWMutex
	}
)

func New(client *elastic.Client) question.Repo {
	r := &repo{
		indexName: "questions",
		client:    client,
		questions: make(map[string]string),
		// mu:        sync.RWMutex{},
	}
	ctx := context.Background()
	if exists, err := client.IndexExists(r.indexName).Do(ctx); err != nil {
		panic(err)
	} else if !exists {
		createdIndex, err := client.CreateIndex(r.indexName).BodyString(indexSettings).Do(ctx)
		if err != nil {
			panic(err)
		} else if !createdIndex.Acknowledged {
			panic("index creation wasn't acknowledged")
		}
	}
	return r
}

func (r *repo) AddQuestion(input *dto.QuestionIn) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.questions[input.UserID] = input.Question
	return nil
}

func (r *repo) GetQuestion(userID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	q, ok := r.questions[userID]
	if !ok {
		return "", domain.EntityNotFound
	}
	return q, nil
}

func (r *repo) DeleteQuestion(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.questions[userID]; !ok {
		return domain.EntityNotFound
	}
	delete(r.questions, userID)
	return nil
}

func (r *repo) Create(in *entity.Question) error {
	ctx := context.Background()
	_, err := r.client.
		Index().
		Index(r.indexName).
		BodyJson(in).
		Do(ctx)

	return err
}

func (r *repo) Update(in *entity.Question) error {
	ctx := context.Background()
	_, err := r.client.
		Update().
		Index(r.indexName).
		Id(in.ID).
		Doc(in).
		Do(ctx)

	return err
}

func (r *repo) Get(question string) (*entity.Question, error) {
	ctx := context.Background()
	query := elastic.NewMatchQuery("question", question).Fuzziness("AUTO")
	searchResult, err := r.client.Search(r.indexName).Query(query).Do(ctx)
	if err != nil {
		return nil, err
	}
	if searchResult.Hits.TotalHits.Value == 0 {
		return nil, domain.EntityNotFound
	}

	found := searchResult.Hits.Hits[0]
	response := &entity.Question{
		ID: found.Id,
	}
	if err := json.Unmarshal(found.Source, response); err != nil {
		return nil, err
	}

	return response, nil
}
