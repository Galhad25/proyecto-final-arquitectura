package service

import (
	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
)

type LaptopService struct {
	laptopRepo *repository.LaptopRepository
}

func NewLaptopService(laptopRepo *repository.LaptopRepository) *LaptopService {
	return &LaptopService{laptopRepo: laptopRepo}
}

func (s *LaptopService) List(status string) ([]models.Laptop, error) {
	return s.laptopRepo.FindAll(status)
}

func (s *LaptopService) Create(req models.LaptopRequest) (*models.Laptop, error) {
	laptop := &models.Laptop{
		Brand:        req.Brand,
		Model:        req.Model,
		SerialNumber: req.SerialNumber,
	}
	if err := s.laptopRepo.Create(laptop); err != nil {
		return nil, err
	}
	return laptop, nil
}
