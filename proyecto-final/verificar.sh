#!/usr/bin/env bash
# Ejecuta el checklist de la Fase 4 contra http://localhost:8080 y muestra cada resultado.
# Uso:  bash proyecto-final/verificar.sh | tee proyecto-final/salida-verificacion.txt
# Requisito: `docker compose up --build` ya está corriendo.
BASE=${BASE:-http://localhost:8080}
RUN=$(date +%s)
ADMIN_EMAIL="admin$RUN@test.com"; STU_EMAIL="est$RUN@test.com"; STU2_EMAIL="est2$RUN@test.com"
PASS="secreta123"

req() { # metodo ruta [token] [json]
  local m=$1 p=$2 t=$3 d=$4
  local args=(-s -w '\n[HTTP %{http_code}]' -X "$m" "$BASE$p" -H 'Content-Type: application/json')
  [ -n "$t" ] && args+=(-H "Authorization: Bearer $t")
  [ -n "$d" ] && args+=(-d "$d")
  curl "${args[@]}"
}
tok() { echo "$1" | grep -o '"token":"[^"]*"' | head -1 | cut -d'"' -f4; }
idof() { echo "$1" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2; }
h() { echo; echo "=== $1"; }

h "docker compose ps"; docker compose ps
h "GET /health";  req GET /health
h "GET /ping";    req GET /ping

h "POST /api/auth/register (admin candidato)"; R=$(req POST /api/auth/register "" "{\"name\":\"Admin\",\"email\":\"$ADMIN_EMAIL\",\"password\":\"$PASS\"}"); echo "$R"
h "Promover a admin vía SQL"
docker compose exec -T mysql mysql -uroot -proot_pass biblioteca -e "UPDATE users SET role='admin' WHERE email='$ADMIN_EMAIL';" 2>/dev/null && echo "ok"
h "POST /api/auth/login (admin)"; R=$(req POST /api/auth/login "" "{\"email\":\"$ADMIN_EMAIL\",\"password\":\"$PASS\"}"); echo "$R"; AT=$(tok "$R")
h "POST /api/auth/register (estudiante 1)"; R=$(req POST /api/auth/register "" "{\"name\":\"Est Uno\",\"email\":\"$STU_EMAIL\",\"password\":\"$PASS\"}"); echo "$R"; ST=$(tok "$R")
h "POST /api/auth/register (estudiante 2)"; R=$(req POST /api/auth/register "" "{\"name\":\"Est Dos\",\"email\":\"$STU2_EMAIL\",\"password\":\"$PASS\"}"); echo "$R"; ST2=$(tok "$R")
h "POST /api/auth/login (estudiante 1)"; req POST /api/auth/login "" "{\"email\":\"$STU_EMAIL\",\"password\":\"$PASS\"}"

echo; echo "######## LIBROS (nada existente roto) ########"
h "GET /api/books"; req GET /api/books "$ST" | head -c 400; echo
h "POST /api/loans (libro 1)"; R=$(req POST /api/loans "$ST" '{"book_id":1}'); echo "$R"; BL=$(idof "$R")
h "GET /api/loans/me"; req GET /api/loans/me "$ST"
h "PUT /api/loans/$BL/return"; req PUT "/api/loans/$BL/return" "$ST"

echo; echo "######## LAPTOPS (nuevo) ########"
h "POST /api/laptops (admin) -> 201";  R=$(req POST /api/laptops "$AT" "{\"brand\":\"Asus\",\"model\":\"VivoBook\",\"serial_number\":\"ASU-$RUN\"}"); echo "$R"; LID=$(idof "$R")
h "POST /api/laptops serial repetido -> 409"; req POST /api/laptops "$AT" "{\"brand\":\"Asus\",\"model\":\"VivoBook\",\"serial_number\":\"ASU-$RUN\"}"
h "POST /api/laptops como estudiante -> 403"; req POST /api/laptops "$ST" '{"brand":"x","model":"y","serial_number":"z"}'
h "POST /api/laptops sin token -> 401"; req POST /api/laptops "" '{"brand":"x","model":"y","serial_number":"z"}'
h "GET /api/laptops"; req GET /api/laptops "$ST"
h "GET /api/laptops?status=available"; req GET "/api/laptops?status=available" "$ST"
h "POST /api/laptop-loans (prestar) -> 201"; R=$(req POST /api/laptop-loans "$ST" "{\"laptop_id\":$LID}"); echo "$R"; LL=$(idof "$R")
h "Prestar la MISMA laptop con otro usuario -> 409"; req POST /api/laptop-loans "$ST2" "{\"laptop_id\":$LID}"
h "Mismo usuario pide OTRA laptop -> 409"; req POST /api/laptop-loans "$ST" '{"laptop_id":1}'
h "Laptop inexistente -> 404"; req POST /api/laptop-loans "$ST2" '{"laptop_id":99999}'
h "GET /api/laptop-loans/me"; req GET /api/laptop-loans/me "$ST"
h "Devolver con otro usuario -> 403"; req PUT "/api/laptop-loans/$LL/return" "$ST2"
h "PUT /api/laptop-loans/$LL/return -> 200"; req PUT "/api/laptop-loans/$LL/return" "$ST"
h "Devolver otra vez -> 409"; req PUT "/api/laptop-loans/$LL/return" "$ST"
h "Prestar de nuevo la laptop devuelta -> 201"; req POST /api/laptop-loans "$ST2" "{\"laptop_id\":$LID}"
echo; echo "FIN"
