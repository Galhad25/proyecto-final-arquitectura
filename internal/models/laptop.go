package models

import "time"

// Laptop representa una laptop física del inventario de la biblioteca.
type Laptop struct {
	ID           int       `db:"id" json:"id"`
	Brand        string    `db:"brand" json:"brand"`
	Model        string    `db:"model" json:"model"`
	SerialNumber string    `db:"serial_number" json:"serial_number"`
	Status       string    `db:"status" json:"status"` // available | loaned
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

// LaptopRequest es el payload para registrar una laptop en el inventario.
type LaptopRequest struct {
	Brand        string `json:"brand" binding:"required" example:"Lenovo"`
	Model        string `json:"model" binding:"required" example:"ThinkPad E14"`
	SerialNumber string `json:"serial_number" binding:"required" example:"LNV-0002"`
}

// LaptopLoan representa el préstamo de una laptop a un usuario.
type LaptopLoan struct {
	ID         int        `db:"id" json:"id"`
	UserID     int        `db:"user_id" json:"user_id"`
	LaptopID   int        `db:"laptop_id" json:"laptop_id"`
	LoanedAt   time.Time  `db:"loaned_at" json:"loaned_at"`
	DueAt      time.Time  `db:"due_at" json:"due_at"`
	ReturnedAt *time.Time `db:"returned_at" json:"returned_at,omitempty"`
	Status     string     `db:"status" json:"status"`
}

// LaptopLoanRequest es el payload para pedir prestada una laptop.
type LaptopLoanRequest struct {
	LaptopID int `json:"laptop_id" binding:"required" example:"1"`
}

// LaptopLoanDetail combina el préstamo con los datos básicos de la laptop.
type LaptopLoanDetail struct {
	LaptopLoan
	LaptopBrand  string `db:"laptop_brand" json:"laptop_brand"`
	LaptopModel  string `db:"laptop_model" json:"laptop_model"`
	LaptopSerial string `db:"laptop_serial" json:"laptop_serial"`
}
