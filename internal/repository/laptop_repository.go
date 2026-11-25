package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"biblioteca-api/internal/models"
)

// ErrDuplicateSerial se devuelve cuando ya existe una laptop con ese número de serie.
var ErrDuplicateSerial = errors.New("ya existe una laptop con ese número de serie")

type LaptopRepository struct {
	db *sqlx.DB
}

func NewLaptopRepository(db *sqlx.DB) *LaptopRepository {
	return &LaptopRepository{db: db}
}

// FindAll devuelve el inventario, opcionalmente filtrado por estado (available | loaned).
func (r *LaptopRepository) FindAll(status string) ([]models.Laptop, error) {
	query := `SELECT * FROM laptops WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY brand ASC, model ASC, id ASC`

	laptops := []models.Laptop{}
	if err := r.db.Select(&laptops, query, args...); err != nil {
		return nil, err
	}
	return laptops, nil
}

func (r *LaptopRepository) FindByID(id int) (*models.Laptop, error) {
	var l models.Laptop
	if err := r.db.Get(&l, `SELECT * FROM laptops WHERE id = ?`, id); err != nil {
		return nil, err
	}
	return &l, nil
}

func (r *LaptopRepository) Create(l *models.Laptop) error {
	query := `INSERT INTO laptops (brand, model, serial_number, status) VALUES (?, ?, ?, 'available')`
	result, err := r.db.Exec(query, l.Brand, l.Model, l.SerialNumber)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return ErrDuplicateSerial
		}
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	l.ID = int(id)
	l.Status = "available"
	return nil
}

// TryMarkLoaned marca la laptop como prestada SOLO si estaba disponible.
// Es una operación atómica (UPDATE condicional): devuelve true si esta llamada
// logró reservarla y false si otra petición se la llevó antes.
func (r *LaptopRepository) TryMarkLoaned(id int) (bool, error) {
	result, err := r.db.Exec(`UPDATE laptops SET status = 'loaned' WHERE id = ? AND status = 'available'`, id)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}

// MarkAvailable devuelve la laptop al inventario.
func (r *LaptopRepository) MarkAvailable(id int) error {
	_, err := r.db.Exec(`UPDATE laptops SET status = 'available' WHERE id = ?`, id)
	return err
}
