# Fase 5 — Diagramas (Mermaid)

Todos reflejan la decisión del [ADR-001](adr.md): módulo de laptops **paralelo** al de libros, con tablas nuevas `laptops` y `laptop_loans`.

## 1. C4 Nivel 1 — Contexto

```mermaid
C4Context
  title Contexto — Sistema de Biblioteca (libros y laptops)
  Person(student, "Estudiante", "Consulta libros y laptops, y pide prestado")
  Person(admin, "Administrador", "Gestiona el catálogo de libros y el inventario de laptops")
  System(biblioteca, "Biblioteca API", "Gestiona usuarios, libros, laptops y préstamos")
  Rel(student, biblioteca, "Consulta y pide prestado", "HTTPS/JSON")
  Rel(admin, biblioteca, "Administra catálogo e inventario", "HTTPS/JSON")
```

## 2. C4 Nivel 2 — Contenedores

```mermaid
C4Container
  title Contenedores — Biblioteca API
  Person(user, "Usuario", "Estudiante o administrador")
  Container(app, "App cliente", "React / Swagger / curl", "Consume la API con un token JWT")
  System_Boundary(sys, "Docker Compose") {
    Container(api, "API REST", "Go + Gin :8080", "Auth JWT, libros, préstamos de libros, laptops, préstamos de laptops")
    ContainerDb(db, "MySQL 8", "MySQL :3306", "users, books, loans, laptops, laptop_loans")
    Container(adminer, "Adminer", "PHP :8081", "Visor web de la base de datos")
  }
  Rel(user, app, "Usa")
  Rel(app, api, "Peticiones REST", "HTTP/JSON + JWT")
  Rel(api, db, "Lee y escribe", "SQL (sqlx)")
  Rel(adminer, db, "Inspecciona", "SQL")
```

## 3. Diagrama de componentes — API con el módulo de laptops

```mermaid
flowchart LR
  subgraph MW[Middleware]
    AUTH[AuthRequired - JWT]
    ADM[AdminRequired]
  end

  subgraph H[Handlers]
    AH[AuthHandler]
    BH[BookHandler]
    LH[LoanHandler]
    LPH[LaptopHandler]:::new
    LLH[LaptopLoanHandler]:::new
  end

  subgraph S[Services]
    AS[AuthService]
    BS[BookService]
    LS[LoanService]
    LPS[LaptopService]:::new
    LLS[LaptopLoanService]:::new
  end

  subgraph R[Repositories]
    UR[UserRepository]
    BR[BookRepository]
    LR[LoanRepository]
    LPR[LaptopRepository]:::new
    LLR[LaptopLoanRepository]:::new
  end

  DB[(MySQL)]

  AUTH --> AH & BH & LH & LPH & LLH
  ADM --> BH & LPH

  AH --> AS --> UR
  BH --> BS --> BR
  LH --> LS --> LR
  LS --> BR
  LPH --> LPS --> LPR
  LLH --> LLS --> LLR
  LLS --> LPR

  UR & BR & LR & LPR & LLR --> DB

  classDef new fill:#d4f4dd,stroke:#2e7d32,color:#000;
```

*Verde = componentes nuevos. Los módulos de libros y laptops no se llaman entre sí: ese es el efecto de la opción B.*

## 4. Diagrama de secuencia — "Prestar una laptop"

```mermaid
sequenceDiagram
  autonumber
  actor U as Usuario
  participant MW as AuthRequired
  participant H as LaptopLoanHandler
  participant S as LaptopLoanService
  participant LR as LaptopRepository
  participant LLR as LaptopLoanRepository
  participant DB as MySQL

  U->>MW: POST /api/laptop-loans {laptop_id} + Bearer JWT
  MW->>MW: valida token y guarda user_id / role
  MW->>H: Borrow()
  H->>H: ShouldBindJSON (laptop_id requerido → 400)
  H->>S: Borrow(userID, req)
  S->>LR: FindByID(laptop_id)
  LR->>DB: SELECT * FROM laptops WHERE id=?
  Note over S: VALIDACIÓN 1: laptop existe (si no → 404)
  S->>LLR: HasActiveLoan(userID)
  LLR->>DB: SELECT COUNT(*) FROM laptop_loans WHERE user_id=? AND status='active'
  Note over S: VALIDACIÓN 2: el usuario no tiene otra laptop activa (si no → 409)
  S->>LR: TryMarkLoaned(laptop_id)
  LR->>DB: UPDATE laptops SET status='loaned' WHERE id=? AND status='available'
  Note over S: VALIDACIÓN 3 (atómica): RowsAffected = 1 → disponible (si 0 → 409)
  S->>LLR: Create(userID, laptop_id, due_at = hoy + 7 días)
  LLR->>DB: INSERT INTO laptop_loans (...)
  Note over S: Si el INSERT falla → MarkAvailable (compensación)
  S-->>H: LaptopLoan
  H-->>U: 201 Created
```

### Devolución (inventario de vuelta)

```mermaid
sequenceDiagram
  autonumber
  actor U as Usuario
  participant H as LaptopLoanHandler
  participant S as LaptopLoanService
  participant LLR as LaptopLoanRepository
  participant LR as LaptopRepository
  U->>H: PUT /api/laptop-loans/:id/return + JWT
  H->>S: Return(id, userID, role)
  S->>LLR: FindByID(id)  (no existe → 404)
  Note over S: ¿es el dueño o admin? (si no → 403)
  Note over S: ¿ya devuelto? (si sí → 409)
  S->>LLR: MarkReturned(id)
  S->>LR: MarkAvailable(laptop_id)
  H-->>U: 200 OK
```
