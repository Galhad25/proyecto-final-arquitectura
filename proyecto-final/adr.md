# ADR-001 — Modelar el préstamo de laptops como un módulo paralelo al de libros

- **Estado:** Aceptado
- **Fecha:** 25 de noviembre de 2026

## Contexto

La biblioteca empezará a prestar **laptops**, además de libros. Hoy el préstamo está acoplado a "libro" en todas las capas: `loans.book_id` es una FK `NOT NULL` hacia `books`, y `Loan`, `LoanRepository`, `LoanService` y las rutas `/api/loans*` asumen un libro (ver [análisis](analisis.md)). Además, el inventario de libros funciona por **copias** (`available_copies`), mientras que una laptop es un objeto individual.

Restricciones: (1) el módulo de libros, la autenticación, `/health` y `/ping` **deben seguir funcionando exactamente igual**; (2) el plazo del proyecto es corto; (3) hoy existe **un solo** recurso nuevo.

Opciones consideradas:

- **A.** Generalizar `loans` con `resource_type/resource_id`.
- **B.** Dominio paralelo: tablas `laptops` y `laptop_loans` + archivos `laptop_*` por capa.
- **C.** Modelo genérico `resources` + `resource_loans`, migrando libros.

## Decisión

**Elegimos la opción B (dominio paralelo).**

Por qué sobre las otras, según los criterios del curso:

1. **Riesgo (criterio decisivo).** B es puramente aditiva: una migración nueva con tablas nuevas y archivos nuevos. El código de libros no se modifica; solo se *registran* los nuevos handlers en `routes.go` y `main.go`. A modifica `loans` y `LoanService` (los mismos que ya funcionan); C reescribe y migra datos estables. La regla de oro del proyecto —nada existente se rompe— se cumple por construcción con B.
2. **Integridad de datos.** `laptop_loans.laptop_id` conserva una **FK real** hacia `laptops`. En A, un `resource_id` genérico pierde la integridad referencial que hoy garantiza la base de datos.
3. **Cohesión.** El dominio de laptops tiene reglas propias (inventario individual por serial, 7 días, una laptop por usuario) distintas a las de libros (copias, 14 días). Mezclarlas en `LoanService` (opción A) lo llenaría de ramas `if tipo`, violando Open/Closed.
4. **Simplicidad / YAGNI.** C resuelve bien el problema de "N recursos", pero hoy hay 1. Construir una abstracción genérica con un solo caso de uso es especulación: se suele abstraer mal sin ver al menos dos o tres casos reales.

Decisiones de negocio (ADR-002 incluido aquí por brevedad):

| Pregunta | Decisión | Justificación |
|---|---|---|
| ¿Inventario individual o por modelo? | **Individual, por número de serie** (`laptops.serial_number UNIQUE`, `status available/loaned`). | Una laptop es un activo físico trazable (daños, pérdidas, garantía). Con copias por modelo no se sabría *cuál* equipo tiene cada usuario. |
| ¿Días de préstamo? | **7 días** (constante `laptopLoanDurationDays`). | Es un activo escaso y costoso; 14 días como libros la dejaría indisponible demasiado tiempo. |
| ¿Quién puede pedirla? | **Cualquier usuario autenticado** (estudiante o admin); solo el admin crea laptops. | Mismas reglas de auth y roles que libros. |
| ¿Límite por usuario? | **Una laptop activa por usuario.** | Evita acaparar el inventario; es la regla equivalente a "no duplicar préstamo" de libros. |
| ¿Quién devuelve? | **El dueño del préstamo o un admin.** | El módulo de libros no verifica dueño; aquí se corrige para laptops (activo de mayor valor). |

## Consecuencias

✅ **Lo que se gana**
- Cero cambios en tablas y servicios existentes; los libros no pueden romperse por esta decisión.
- Reglas de negocio claras y propias del dominio laptops, fáciles de probar.
- Reserva **atómica** (`UPDATE … WHERE status='available'` + `RowsAffected`): dos peticiones simultáneas no pueden llevarse la misma laptop.
- Entrega rápida y reversible: quitar el módulo es borrar archivos y una migración.

❌ **Lo que se pierde / deuda aceptada**
- **Duplicación** de la forma del módulo de préstamo (repo/service/handler). Si mañana llegan proyectores, habría un tercer módulo copiado.
- No existe un listado unificado "todos los préstamos del usuario" (libros + laptops): requiere dos consultas.
- Mantener dos flujos de préstamo implica dos conjuntos de pruebas.

**Impacto:** Base de datos: 2 tablas nuevas (`laptops`, `laptop_loans`) en `migrations/002_laptops.sql`, sin tocar las existentes. Capas: se agregan archivos `laptop_*` siguiendo el patrón `handler → service → repository`. Futuro: la **regla de tres** marca el punto de refactor — con el segundo recurso nuevo (por ejemplo, proyectores) se evaluará migrar a la opción C, ya con casos reales que informen la abstracción. Esta decisión no cierra esa puerta: C sigue siendo posible después.

## Si mañana prestan proyectores

Se repetiría el patrón B (rápido, bajo riesgo) o, preferiblemente, ese sería el momento de extraer un módulo genérico de recursos con dos casos reales (laptops y proyectores) y migrar libros de forma incremental.

## Nota sobre un hallazgo en el código existente

`migrations/001_init.sql` tenía un `;` de más antes de la última fila del `INSERT` de libros (error de sintaxis que aborta la inicialización de MySQL con un volumen nuevo). Se corrigió en un commit **separado** (`fix(migrations)`), sin cambiar datos ni esquema.
