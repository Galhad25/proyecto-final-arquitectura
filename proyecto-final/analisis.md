# Fase 1 — Análisis del problema

> Basado en la lectura del código real de `biblioteca-api` (no en suposiciones).

## 1. ¿Dónde está acoplado el préstamo a "libro"?

| Capa | Archivo | Punto de acoplamiento |
|---|---|---|
| Base de datos | `migrations/001_init.sql` | `loans.book_id INT NOT NULL` con `FOREIGN KEY → books(id)`: una fila de `loans` **solo puede apuntar a un libro**. |
| Modelo | `internal/models/loan.go` | `Loan.BookID`, `LoanRequest.BookID` y `LoanDetail.BookTitle/BookAuthor` (el DTO de lectura asume datos de libro). |
| Repository | `internal/repository/loan_repository.go` | `Create(userID, bookID, …)`, `HasActiveLoan(userID, bookID)` y los `JOIN books b ON b.id = l.book_id` en `FindByUser` y `FindAll`. |
| Service | `internal/service/loan_service.go` | `Borrow()` depende de `BookRepository`: valida `AvailableCopies`, rechaza duplicados por libro, usa `loanDurationDays = 14` y llama a `DecrementAvailable/IncrementAvailable`. Los errores (`ErrBookNotAvailable`…) también hablan de libros. |
| Handler / rutas | `internal/handler/loan_handler.go`, `routes.go` | Payload `book_id`, mensajes "libro devuelto correctamente", rutas `/api/loans*`. |

**Conclusión:** el concepto "préstamo" y el concepto "libro" están fusionados en las 5 capas. Además, el modelo de inventario del libro es *por copias* (`total_copies/available_copies`), y eso **no encaja de forma natural** con una laptop, que es un objeto físico identificable (serial).

## 2. ¿Qué pasaría si mañana prestaran proyectores, salas o tablets?

Con el diseño actual, cada recurso nuevo exige repetir el trabajo completo: tabla + modelo + repo + servicio + handler + rutas. Y si se intentara "reutilizar" `loans`, habría que romper la FK `NOT NULL` hacia `books`, lo cual toca la tabla más sensible del sistema. Hoy el costo de agregar un recurso es **alto y repetitivo**; el diseño no es extensible por construcción (atributo de calidad de la semana 10).

## 3. Opciones de diseño

### Opción A — Generalizar la tabla `loans` (polimórfica)
Agregar `resource_type` + `resource_id` (o `laptop_id` nullable) a `loans`, quitar el `NOT NULL`/FK de `book_id` y adaptar `LoanService` para atender ambos tipos.

### Opción B — Dominio paralelo (tablas y módulo nuevos)
Crear `laptops` y `laptop_loans` y archivos nuevos `laptop_*` en cada capa, sin modificar el módulo de libros. Solo se tocan `routes.go` y `main.go` para **registrar** lo nuevo.

### Opción C — Modelo genérico de "recursos"
Crear `resources` (con `type`: book, laptop…) y `resource_loans`; migrar los libros y préstamos existentes a esa estructura y reescribir `BookService`/`LoanService` encima.

## 4. Tabla comparativa de trade-offs

| Opción | Qué modifica de la BD existente | Qué duplica / qué reutiliza | Costo de agregar otro recurso (proyectores) | Riesgo sobre lo que ya funciona |
|---|---|---|---|---|
| **A. `loans` polimórfica** | **Altera `loans`** (quita NOT NULL/FK de `book_id`, agrega columnas). Pierde integridad referencial (con `resource_id` genérico no hay FK real). | Reutiliza tabla y servicio de préstamos; el servicio se llena de `if tipo == …`. | Medio: una columna/valor más, pero cada tipo suma ramas al mismo servicio (violación Open/Closed). | **Alto:** modifica la tabla y el servicio en producción; un error rompe préstamos de libros. |
| **B. Dominio paralelo** | **Nada**: solo agrega tablas nuevas (`002_laptops.sql`). | Duplica la *forma* del módulo de préstamo (repo/service/handler); reutiliza `users`, JWT, middleware y el patrón de capas. | Medio-alto: otro módulo copiado por cada recurso. Se vuelve deuda si llegan 3+ recursos. | **Mínimo:** el código de libros no se toca (aditivo). |
| **C. `resources` genérico** | **Migra todo**: libros y préstamos existentes pasan a tablas nuevas. | Casi nada duplicado; una sola lógica de préstamo. | **Bajo:** un recurso nuevo = un `type` y sus atributos. | **Muy alto:** migración de datos + reescritura de módulos estables; mucho trabajo para un solo recurso nuevo. |

## 5. Lectura del análisis

- **Acoplamiento y cohesión (sem. 3):** A reduce la cohesión del servicio de préstamos; B mantiene módulos cohesivos pero con duplicación; C es el más cohesivo a largo plazo.
- **Extensibilidad (sem. 10):** C > A > B en el largo plazo; B es la más barata *hoy*.
- **Riesgo / regla de oro del proyecto ("nada existente se rompe"):** B < A < C.
- Lo que hay hoy es **un solo recurso nuevo**, no una familia de recursos: generalizar ahora sería diseñar para un futuro que aún no existe (YAGNI). La decisión se justifica en el [ADR](adr.md).
