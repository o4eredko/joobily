package dto

import validation "github.com/go-ozzo/ozzo-validation/v4"

type (
	QuestionIn struct {
		UserID   string
		Question string
	}
)

func (q *QuestionIn) Validate() error {
	return validation.ValidateStruct(
		q,
		validation.Field(&q.UserID, validation.Required),
		validation.Field(&q.Question, validation.Required),
	)
}
