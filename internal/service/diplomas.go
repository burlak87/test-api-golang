package service

import (
	"errors"
	"fmt"
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

func (d *Diplomas) CreateResource(diploma domain.Diploma) (domain.Diploma, error) { // меняем возвращаемое значение
    if diploma.Title == "" {
        return domain.Diploma{}, errors.New("Title invalid")
    }
    if len(diploma.Description) > 500 {
        return domain.Diploma{}, errors.New("Description too long")
    }

    fmt.Printf("DEBUG SERVICE DIPLOMA CREATE: Calling storage.InsertResource\n")
    id, err := d.storage.InsertResource(diploma)
    if err != nil {
        fmt.Printf("DEBUG SERVICE DIPLOMA CREATE: Storage error: %v\n", err)
        return domain.Diploma{}, err
    }

    createdDiploma := domain.Diploma{
        ID:          id,
        Title:       diploma.Title,
        Description: diploma.Description,
    }
    
    fmt.Printf("DEBUG SERVICE DIPLOMA CREATE: SUCCESS - Created diploma with ID: %d\n", id)
    return createdDiploma, nil
}

func (d *Diplomas) UpdateResource(id int64, diploma domain.Diploma) (domain.Diploma, error) {
    if id == 0 {
        return domain.Diploma{}, errors.New("id invalid")
    }
    if diploma.Title == "" {
        return domain.Diploma{}, errors.New("Title invalid")
    }
    if len(diploma.Description) > 500 {
        return domain.Diploma{}, errors.New("Description too long")
    }

    fmt.Printf("DEBUG SERVICE DIPLOMA UPDATE: Calling storage.RenovationResource\n")
    updatedDiploma, err := d.storage.RenovationResource(id, diploma)
    if err != nil {
        fmt.Printf("DEBUG SERVICE DIPLOMA UPDATE: Storage error: %v\n", err)
        return domain.Diploma{}, err
    }
    
    updatedDiploma.ID = id
    
    fmt.Printf("DEBUG SERVICE DIPLOMA UPDATE: SUCCESS - Updated diploma with ID: %d\n", id)
    return updatedDiploma, nil
}

func (d *Diplomas) DeleteResource(id int64) error {
	err := d.storage.DestroyResource(id)
	if err != nil {
		return err
	}

	return nil
}