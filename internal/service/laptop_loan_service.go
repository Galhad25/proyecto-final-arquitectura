package service

import (
	"database/sql"
	"errors"
	"time"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
)

var (
	ErrLaptopNotFound      = errors.New("laptop no encontrada")
	ErrLaptopNotAvailable  = errors.New("la laptop no está disponible")
	ErrUserHasActiveLaptop = errors.New("ya tienes una laptop prestada; devuélvela antes de pedir otra")
	ErrLaptopLoanNotFound  = errors.New("préstamo de laptop no encontrado")
	ErrLaptopLoanReturned  = errors.New("este préstamo de laptop ya fue devuelto")
	ErrLaptopLoanNotYours  = errors.New("no puedes devolver el préstamo de otro usuario")
)

// Regla de negocio (ADR-002): las laptops se prestan por 7 días.
const laptopLoanDurationDays = 7

type LaptopLoanService struct {
	loanRepo   *repository.LaptopLoanRepository
	laptopRepo *repository.LaptopRepository
}

func NewLaptopLoanService(loanRepo *repository.LaptopLoanRepository, laptopRepo *repository.LaptopRepository) *LaptopLoanService {
	return &LaptopLoanService{loanRepo: loanRepo, laptopRepo: laptopRepo}
}

func (s *LaptopLoanService) Borrow(userID int, req models.LaptopLoanRequest) (*models.LaptopLoan, error) {
	if _, err := s.laptopRepo.FindByID(req.LaptopID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLaptopNotFound
		}
		return nil, err
	}

	hasActive, err := s.loanRepo.HasActiveLoan(userID)
	if err != nil {
		return nil, err
	}
	if hasActive {
		return nil, ErrUserHasActiveLaptop
	}

	// Reserva atómica: solo una petición puede pasar de 'available' a 'loaned'.
	claimed, err := s.laptopRepo.TryMarkLoaned(req.LaptopID)
	if err != nil {
		return nil, err
	}
	if !claimed {
		return nil, ErrLaptopNotAvailable
	}

	dueAt := time.Now().AddDate(0, 0, laptopLoanDurationDays)
	loan, err := s.loanRepo.Create(userID, req.LaptopID, dueAt)
	if err != nil {
		// Compensación: si no se pudo registrar el préstamo, liberamos la laptop.
		_ = s.laptopRepo.MarkAvailable(req.LaptopID)
		return nil, err
	}
	return loan, nil
}

// Return cierra el préstamo. Solo lo puede devolver su dueño o un admin.
func (s *LaptopLoanService) Return(loanID, userID int, role string) error {
	loan, err := s.loanRepo.FindByID(loanID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrLaptopLoanNotFound
		}
		return err
	}
	if loan.UserID != userID && role != "admin" {
		return ErrLaptopLoanNotYours
	}
	if loan.Status == "returned" {
		return ErrLaptopLoanReturned
	}

	if err := s.loanRepo.MarkReturned(loanID); err != nil {
		return err
	}
	return s.laptopRepo.MarkAvailable(loan.LaptopID)
}

func (s *LaptopLoanService) ListByUser(userID int) ([]models.LaptopLoanDetail, error) {
	return s.loanRepo.FindByUser(userID)
}
