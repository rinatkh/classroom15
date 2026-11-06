# Полный flow через curl

Для автоматической демонстрации с проверкой всех statuses: запустить финальный API
с ACCESS_TOKEN_TTL=20s, COOKIE_SECURE=false и выполнить `make session-demo`.
Script использует cookie jars `curl -b/-c`, временные файлы удаляет, tokens не печатает.

Для ручного demo (в пределах своего terminal, не включать `set -x`):

```bash
umask 077
DEMO_DIR="$(mktemp -d)"
BASE_URL=http://localhost:8080
curl -sS -X POST "$BASE_URL/register" -H 'Content-Type: application/json' \
  -d '{"email":"alice-manual@example.test","password":"secret123"}'
curl -sS -c "$DEMO_DIR/current.jar" -o "$DEMO_DIR/login.json" \
  -X POST "$BASE_URL/login" -H 'Content-Type: application/json' \
  -d '{"email":"alice-manual@example.test","password":"secret123"}'
ACCESS="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])' < "$DEMO_DIR/login.json")"
curl -sS "$BASE_URL/tasks" -H "Authorization: Bearer $ACCESS"
cp "$DEMO_DIR/current.jar" "$DEMO_DIR/old.jar"
# Wait until this access JWT expires (20s only in demo config).
sleep 21
curl -sS -o /dev/null -w '%{http_code}\n' "$BASE_URL/tasks" -H "Authorization: Bearer $ACCESS"
# 401; browser would automatically attach the HttpOnly cookie here:
curl -sS -b "$DEMO_DIR/current.jar" -c "$DEMO_DIR/current.jar" \
  -X POST "$BASE_URL/auth/refresh" -o "$DEMO_DIR/refresh.json"
ACCESS="$(python3 -c 'import json,sys; print(json.load(sys.stdin)["access_token"])' < "$DEMO_DIR/refresh.json")"
curl -sS "$BASE_URL/tasks" -H "Authorization: Bearer $ACCESS"
curl -sS -o /dev/null -w '%{http_code}\n' -b "$DEMO_DIR/old.jar" -X POST "$BASE_URL/auth/refresh"
# 401: old credential revoked by rotation.
curl -sS -o /dev/null -w '%{http_code}\n' -b "$DEMO_DIR/current.jar" -c "$DEMO_DIR/current.jar" -X POST "$BASE_URL/auth/logout"
# 204. Cookie removed. Access still valid until exp.
curl -sS -o /dev/null -w '%{http_code}\n' -b "$DEMO_DIR/current.jar" -X POST "$BASE_URL/auth/refresh"
# 401; new login needed.
rm -rf "$DEMO_DIR"
unset ACCESS
```

Повторный запуск register с тем же email ожидаемо 409: выберите новый адрес или login.
Для restart demonstration сделайте отдельный login и сохраните current.jar.
Перезапустите API с тем же DSN/JWT_SECRET, до logout вызовите refresh с этим jar:
200. Не используйте jar, отозванный session-demo. Проверка DB state без секретов:
`psql "$DATABASE_URL" -f sql/10_refresh_sessions.sql`.
В браузерном JS cookie читает/прикладывает сам браузер, `curl` только моделирует его jar.
