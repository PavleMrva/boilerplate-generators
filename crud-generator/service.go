package yourmodel

// Code generated by "crud-generator"; Feel free to make changes in order to align service to your business logic.

import (
	"context"
)

type Service interface {
	Add(ctx context.Context, yourModel *YourModel) error
	Get(ctx context.Context, id uint) (*YourModel, error)
	Edit(ctx context.Context, yourModel *YourModel) error
	Remove(ctx context.Context, id uint) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Add(ctx context.Context, yourModel *YourModel) error {
	err = s.repository.Insert(ctx, yourModel)
	if err != nil {
		return errors.Wrap(err, "repository insert")
	}

	return nil
}

func (s *service) Get(ctx context.Context, id uint) (*YourModel, error) {
	res, err := s.repository.Find(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "repository find")
	}

	return res, nil
}

func (s *service) Edit(ctx context.Context, yourModel *YourModel) error {
	_, err := s.repository.Find(ctx, yourModel.ID)
	if err != nil {
		return errors.Wrap(err, "repository find")
	}
	
	if err = s.repository.Update(ctx, yourModel); err != nil {
		return errors.Wrap(err, "repository update")
	}

	return nil
}

func (s *service) Remove(ctx context.Context, id uint) error {
	_, err := s.repository.Find(ctx, id)
	if err != nil {
		return errors.Wrap(err, "repository find")
	}

	if err = s.repository.Delete(ctx, yourModel); err != nil {
		return errors.Wrap(err, "repository delete")
	}
}