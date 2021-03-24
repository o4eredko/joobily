package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	Question struct {
		ID       string `json:"-"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
)

func (q *Question) Validate() error {
	return validation.ValidateStruct(
		q,
		validation.Field(&q.Question, validation.Required),
		validation.Field(&q.Answer, validation.Required),
	)
}
