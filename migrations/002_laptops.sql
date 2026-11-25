-- Módulo de préstamo de laptops (dominio paralelo al de libros, ver ADR-001).
-- No modifica ninguna tabla existente: solo agrega tablas nuevas.

SET NAMES utf8mb4;

-- Cada fila es UNA laptop física, identificada por su número de serie (ADR-002).
CREATE TABLE IF NOT EXISTS laptops (
  id INT AUTO_INCREMENT PRIMARY KEY,
  brand VARCHAR(80) NOT NULL,
  model VARCHAR(120) NOT NULL,
  serial_number VARCHAR(80) NOT NULL UNIQUE,
  status ENUM('available', 'loaned') NOT NULL DEFAULT 'available',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS laptop_loans (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id INT NOT NULL,
  laptop_id INT NOT NULL,
  loaned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  due_at TIMESTAMP NOT NULL,
  returned_at TIMESTAMP NULL,
  status ENUM('active', 'returned') NOT NULL DEFAULT 'active',
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (laptop_id) REFERENCES laptops(id)
);

-- Datos de ejemplo
INSERT INTO laptops (brand, model, serial_number) VALUES
('Lenovo', 'ThinkPad E14', 'LNV-0001'),
('Dell', 'Latitude 3420', 'DEL-0001'),
('HP', 'ProBook 450 G9', 'HP-0001');
