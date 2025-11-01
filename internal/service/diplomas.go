package service

import (
	"errors"
	"gosmol/internal/domain"
)

type DiplomasStorage interface {
	SelectAllResource(limits int64, page int64) ([]domain.Diploma, error)
	SelectResource(id int64) (domain.Diploma, error)
	InsertResource(diploma domain.Diploma) (int64, error)
	RenovationResource(id int64, diploma domain.Diploma) (domain.Diploma, error)
	DestroyResource(id int64) error
}

type Diplomas struct {
	storage DiplomasStorage
}

func NewDiplomas(storage DiplomasStorage) *Diplomas{
	return &Diplomas{storage: storage}
}

func (d *Diplomas) GetResources(limits int64) ([]domain.Diploma, error) {
	var page int64 = 10
	diplomas, err := d.storage.SelectAllResource(limits, page)
	if err != nil {
		return nil, err
	}

	return diplomas, nil
}

func (d *Diplomas) GetResource(id int64) (domain.Diploma, error) {
	diploma, err := d.storage.SelectResource(id)
	if err != nil {
		return domain.Diploma{}, err
	}

	return diploma, nil
}

func (d *Diplomas) CreateResource(diploma domain.Diploma) error {
	if diploma.Title == "" {
		return errors.New("Title invalid")
	}
	if len(diploma.Description) > 500 {
		return errors.New("Description too long")
	}

	_, err := d.storage.InsertResource(diploma)
	if err != nil {
		return err
	}

	return nil
}

func (d *Diplomas) UpdateResource(id int64, diploma domain.Diploma) error {
	if id == 0 {
		return errors.New("id invalid")
	}
	if diploma.Title == "" {
		return errors.New("Title invalid")
	}
	if len(diploma.Description) > 500 {
		return errors.New("Description too long")
	}

	_, err := d.storage.RenovationResource(id, diploma)
	if err != nil {
		return err
	}

	return nil
}

func (d *Diplomas) DeleteResource(id int64) error {
	err := d.storage.DestroyResource(id)
	if err != nil {
		return err
	}

	return nil
}