package person

import (
	"github.com/aarongreenlee/eventsource"
	"github.com/aarongreenlee/eventsource/repository"
)

func NewService(repoOpts ...repository.Option) (*Service, error) {

	repo, err := repository.New(
		&Person{},
		[]eventsource.Event{
			CreateEvent{},
		},
		repoOpts...,
	)
	if err != nil {
		return nil, err
	}

	return &Service{repo: repo}, nil
}

type Service struct {
	repo *repository.Repository
}
