package person

import (
	"context"
	"fmt"

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

func (s Service) Load(ctx context.Context, aggregateID string) (Person, error) {
	aggregate, err := s.repo.Load(ctx, aggregateID)
	if err != nil {
		return Person{}, err
	}

	p, ok := aggregate.(*Person)
	if !ok {
		return Person{}, fmt.Errorf("unable to cast aggregate to %T", Person{})
	}

	return *p, nil
}
