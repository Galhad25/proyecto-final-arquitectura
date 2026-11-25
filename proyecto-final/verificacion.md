# Fase 4 — Verificación

Cómo se generó la evidencia:

```bash
docker compose down -v          # volumen limpio para que corran 001 y 002
docker compose up --build -d
bash proyecto-final/verificar.sh | tee proyecto-final/salida-verificacion.txt
```



## Checklist

| # | Punto | Resultado esperado | Evidencia |
|---|---|---|---|
| 1 | `docker compose up --build` sin errores | build y arranque OK | _pegar salida_ |
| 2 | `docker compose ps` | 3 contenedores `Up`, MySQL `(healthy)` | _pegar salida_ |
| 3 | `GET /health` | `{"status":"ok"}` | _pegar salida_ |
| 4 | `GET /ping` | `{"message":"pong"}` | _pegar salida_ |
| 5 | `POST /api/auth/register` y `/login` | 201/200 con token | _pegar salida_ |
| 6 | `GET /api/books` (con token) | catálogo de libros, 200 | _pegar salida_ |
| 7 | Préstamo de libros: `POST /api/loans`, `GET /api/loans/me`, `PUT /api/loans/:id/return` | 201 / 200 / 200 | _pegar salida_ |
| 8 | `POST /api/laptops` (admin) | 201 con la laptop | _pegar salida_ |
| 9 | `GET /api/laptops` (autenticado) | lista de inventario, 200 | _pegar salida_ |
| 10 | Prestar y devolver una laptop | 201 / 200 | _pegar salida_ |
| 11 | `GET /api/laptop-loans/me` | préstamos del usuario, 200 | _pegar salida_ |
| 12 | Una laptop prestada no se puede prestar dos veces | 409 | _pegar salida_ |

## Casos de error verificados

| Caso | Estado HTTP esperado |
|---|---|
| Crear laptop sin token | 401 |
| Crear laptop como estudiante | 403 |
| Crear laptop con serial repetido | 409 |
| Prestar laptop ya prestada | 409 |
| Pedir una segunda laptop con una activa | 409 |
| Prestar laptop inexistente | 404 |
| Devolver el préstamo de otro usuario | 403 |
| Devolver un préstamo ya devuelto | 409 |

## Qué NO se tocó (garantía de "nada existente se rompe")

`git diff` contra el repo original sobre archivos existentes:

- `internal/handler/routes.go` y `cmd/api/main.go`: solo líneas **agregadas** para registrar los componentes nuevos.
- `migrations/001_init.sql`: una coma en lugar de `;` (bug de sintaxis), commit separado.
- Módulo de libros, autenticación, `/health` y `/ping`: **sin cambios**.
