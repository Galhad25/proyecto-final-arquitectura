package repository

import (
	"time"

	"github.com/jmoiron/sqlx"

	"biblioteca-api/internal/models"
)

type LaptopLoanRepository struct {
	db *sqlx.DB
}

func NewLaptopLoanRepository(db *sqlx.DB) *LaptopLoanRepository {
	return &LaptopLoanRepository{db: db}
}

func (r *LaptopLoanRepository) Create(userID, laptopID int, dueAt time.Time) (*models.LaptopLoan, error) {
	query := `INSERT INTO laptop_loans (user_id, laptop_id, due_at, status) VALUES (?, ?, ?, 'active')`
	result, err := r.db.Exec(query, userID, laptopID, dueAt)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.FindByID(int(id))
}

func (r *LaptopLoanRepository) FindByID(id int) (*models.LaptopLoan, error) {
	var l models.LaptopLoan
	if err := r.db.Get(&l, `SELECT * FROM laptop_loans WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &l, nil
}

// FindByUser devuelve los préstamos de laptop de un usuario, con los datos de la laptop.
func (r *LaptopLoanRepository) FindByUser(userID int) ([]models.LaptopLoanDetail, error) {
	query := `
		SELECT ll.*, l.brand AS laptop_brand, l.model AS laptop_model, l.serial_number AS laptop_serial
		FROM laptop_loans ll
		JOIN laptops l ON l.id = ll.laptop_id
		WHERE ll.user_id = ?
		ORDER BY ll.loaned_at DESC`

	loans := []models.LaptopLoanDetail{}
	if err := r.db.Select(&loans, query, userID); err != nil {
		return nil, err
	}
	return loans, nil
}

func (r *LaptopLoanRepository) MarkReturned(id int) error {
	_, err := r.db.Exec(`UPDATE laptop_loans SET status = 'returned', returned_at = NOW() WHERE id = ?`, id)
	return err
}

// HasActiveLoan indica si el usuario ya tiene una laptop prestada sin devolver.
func (r *LaptopLoanRepository) HasActiveLoan(userID int) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM laptop_loans WHERE user_id = ? AND status = 'active'`
	if err := r.db.Get(&count, query, userID); err != nil {
		return false, err
	}
	return count > 0, nil
}
