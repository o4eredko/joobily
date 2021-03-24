package question

import (
	"gitlab.jooble.com/marketing_tech/joobily/domain"
	"gitlab.jooble.com/marketing_tech/joobily/domain/dto"
	"gitlab.jooble.com/marketing_tech/joobily/domain/entity"
	"gitlab.jooble.com/marketing_tech/joobily/domain/reply"
)

type (
	useCase struct {
		questionRepo     Repo
		pendingQuestions map[string]string // map user id to question
	}

	UseCase interface {
		Get(question string) (string, error)
		Upsert(in *entity.Question) (string, error)
		AddQuestion(userID string, question string) (string, error)
		GetQuestion(userID string) (string, error)
		DiscardQuestion(userID string) (string, error)
		AddAnswer(userID string, answer string) (string, error)
	}

	Repo interface {
		AddQuestion(input *dto.QuestionIn) error
		GetQuestion(userID string) (string, error)
		DeleteQuestion(userID string) error
		Get(question string) (*entity.Question, error)
		Create(in *entity.Question) error
		Update(in *entity.Question) error
	}
)

func New(questionRepo Repo) UseCase {
	return &useCase{
		questionRepo:     questionRepo,
		pendingQuestions: make(map[string]string),
	}
}

func (u *useCase) Get(question string) (string, error) {
	res := reply.InternalError
	q, err := u.questionRepo.Get(question)
	if err != nil {
		if err == domain.EntityNotFound {
			res = reply.AnswerNotFound
		}
		return res, err
	}
	return q.Answer, err
}

func (u *useCase) GetQuestion(userID string) (string, error) {
	return u.questionRepo.GetQuestion(userID)
}

func (u *useCase) AddQuestion(userID string, question string) (string, error) {
	input := &dto.QuestionIn{
		UserID:   userID,
		Question: question,
	}
	if err := input.Validate(); err != nil {
		return reply.ValidationError, err
	}
	if err := u.questionRepo.AddQuestion(input); err != nil {
		return reply.InternalError, err
	}
	return reply.QuestionAddedSuccessfully, nil
}

func (u *useCase) DiscardQuestion(userID string) (string, error) {
	res := reply.QuestionDiscardedSuccessfully
	err := u.questionRepo.DeleteQuestion(userID)
	if err != nil {
		if err == domain.EntityNotFound {
			res = reply.QuestionNotFound
		} else {
			res = reply.InternalError
		}
	}
	return res, err
}

func (u *useCase) AddAnswer(userID string, answer string) (string, error) {
	question, err := u.questionRepo.GetQuestion(userID)
	if err != nil {
		return reply.InternalError, err
	}
	if err := u.questionRepo.DeleteQuestion(userID); err != nil {
		return reply.InternalError, err
	}

	input := &entity.Question{
		Question: question,
		Answer:   answer,
	}
	return u.Upsert(input)
}

func (u *useCase) Upsert(in *entity.Question) (string, error) {
	if err := in.Validate(); err != nil {
		return reply.ValidationError, err
	}
	var m func(*entity.Question) error
	if question, err := u.questionRepo.Get(in.Question); err != nil {
		if err != domain.EntityNotFound {
			return reply.InternalError, err
		}
		m = u.questionRepo.Create
	} else {
		in.ID = question.ID
		m = u.questionRepo.Update
	}
	if err := m(in); err != nil {
		return reply.InternalError, err
	}
	return reply.AnswerAddedSuccessfully, nil
}
