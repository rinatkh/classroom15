# PostgreSQL 16, goose и запуск

Нужны Go 1.22.x, PostgreSQL 16 server/client, psql, Make, Bash, curl, Python 3.
Для race требуется C compiler. Существующая БД занятия 14 уже содержит goose
history: сохраняются DATABASE_URL и JWT_SECRET, применяется только новая 002.

Для новой локальной установки администратор PostgreSQL создаёт отдельные БД:

```sql
CREATE ROLE course_user LOGIN PASSWORD 'course-local-password';
CREATE DATABASE go_course_tasks OWNER course_user;
CREATE DATABASE go_course_tasks_test OWNER course_user;
```

Это локальный пример; пароль и DSN не коммитят. В терминале classroom15:

```bash
export DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks?sslmode=disable'
export TEST_DATABASE_URL='postgres://course_user:course-local-password@localhost:5432/go_course_tasks_test?sslmode=disable'
# Только для новой установки; при upgrade/restart сохраните прежний секрет.
export JWT_SECRET="$(python3 -c 'import secrets; print(secrets.token_hex(32))')"
export ACCESS_TOKEN_TTL=15m
export REFRESH_TOKEN_TTL=168h
export COOKIE_SECURE=false
export LOG_LEVEL=INFO
make status
make up
make status
make structure
make ci
make test-db
make test-db-race
make run
```

На втором терминале `make curl-demo`. Для полного expiration flow перезапустите
API с `ACCESS_TOKEN_TTL=20s`, затем `make session-demo`. По умолчанию Secure=true;
false нужен только для локального HTTP. `.env` автоматически не загружается.

Один `make down` после версии 2 удаляет refresh_sessions; users/tasks остаются.
Ещё один отдельный `make down` удалит базовую схему. Откат удаляет данные:
обратный Up создаст пустую таблицу. После отката 002 API15 нельзя использовать
для login/refresh до повторного `make up`.

`structure.sql` получен командой `make structure` из реальной БД после 001+002,
содержит схему без пользовательских данных, owners/ACL. При восстановлении
не должно быть уже существующей public-схемы с объектами. Goose остаётся
источником версий, снимок — справочник полной структуры.

Troubleshooting: connection refused — проверить server/port; нет таблицы sessions —
проверить `make status` и применить 002; cookie не отправляется — проверить Secure
и HTTPS, имя и Path; 401 после restart — проверить прежний JWT_SECRET; 403 Task —
проверить владельца, refresh не даёт дополнительных прав. DB tests используют
только свои случайные схемы в отдельной TEST_DATABASE_URL, superuser не нужен.
