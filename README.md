# Занятие 15. Access / Refresh tokens, HttpOnly cookies и сессии в PostgreSQL

**Начинаем с финального Task API занятия 14.** Исходный проект остаётся отдельным
репозиторием `classroom14`; внутри classroom15 исторические snapshots не нужны.
Здесь один самостоятельный Go-модуль с `cmd/`, `internal/`, `migrations/`, `sql/`,
`scripts/`, `docs/`, `structure.sql` и Makefile. Homework находится в отдельном
репозитории `homework_go_15` и не входит в classroom.

После 14 уже работают User, Register/Login, bcrypt, JWT, Bearer middleware,
request context, DTO/Mapper, Task UseCase с ownership и PostgreSQL repositories.
Сегодня решаем одну проблему: короткий access JWT истекает, а повторный ввод
пароля каждые 15 минут неудобен. Развиваем authentication поверх того же приложения.

## Что добавлено к занятию 14

В занятии 14: `POST /login` проверяет пароль и возвращает `{"token":"<JWT>"}`.
После `exp` нужен новый login. Обычный protected request проходит через
AuthMiddleware → Task Handler → Task UseCase → Task Repository → PostgreSQL.

В финале Login UseCase дополнительно создаёт refresh session. HTTP Handler
возвращает короткий access в JSON и raw refresh через `Set-Cookie`.
Новый `RefreshSessionRepository` хранит hash и атомарно ротирует сессию.
Обычный Task request проходит по прежней цепочке. Проверка access JWT по подписи
и `exp` не читает refresh_sessions; обработка самих задач по-прежнему использует БД.

## Access и refresh

| | Access token | Refresh token |
|---|---|---|
| Назначение | Обычные API requests | Только refresh/logout |
| Формат | Прежний HS256 JWT: user_id, iat, exp | 32 случайных байта, base64url без padding |
| Передача | Authorization: Bearer | HttpOnly cookie |
| Клиент SPA | Память приложения | Cookie jar браузера |
| Сервер | Проверка подписи и срока | Hash + состояние в PostgreSQL |
| Учебный TTL | 15 минут | 7 дней |

TTL зависят от риска продукта. Это примеры конфигурации, а не стандарт.
Access JWT в БД не сохраняется. Refresh является stateful credential.
Пароль по-прежнему хешируется bcrypt. Высокоэнтропийный случайный refresh
хешируется SHA-256: получаем детерминированный ключ для lookup.

## Схема и atomic rotation

`001_init` из занятия 14 не изменён: users и tasks. Новая migration `002`:

```sql
refresh_sessions (
  id bigint identity primary key,
  user_id bigint not null references users(id) on delete cascade,
  token_hash text not null unique,
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),
  revoked_at timestamptz,
  replaced_by_session_id bigint references refresh_sessions(id)
)
```

`UNIQUE(token_hash)` даёт индекс для lookup. Индекс user_id поддерживает FK
cascade и возможное будущее завершение всех сессий пользователя.
Запись содержит identity, lifetime и revocation, поэтому называется session.

`Refresh` сначала проверяет session на application level. Затем `Rotate` внутри
repository выполняет transaction: SELECT FOR UPDATE, повторная проверка active
и expiration, UPDATE old.revoked_at, INSERT new, связь replaced_by, COMMIT.
При ошибке — ROLLBACK. Под блокировкой проверяем `clock_timestamp()` после
ожидания, потому что `now()` фиксируется в начале transaction.
Из двух одновременных попыток использовать старый refresh успешна ровно одна.
Вторая получает ту же публичную 401, что unknown/expired/revoked refresh.

## Cookie и клиент

`refresh_token=<raw>; HttpOnly; Secure; SameSite=Lax; Path=/auth; Max-Age=...`.
`Secure` включён по умолчанию; для локального HTTP явно задаём false.
Cookie host-only: Domain не задаём. Path ограничивает отправку на `/auth`,
поэтому она не едет на `/tasks`. Это ограничение доставки, а не граница безопасности.
Login остаётся `/login`: ответ может установить cookie с Path=/auth.

HttpOnly запрещает чтение через `document.cookie`, но **не предотвращает CSRF**
и не делает XSS безвредным. SameSite=Lax уменьшает cross-site CSRF risk.
Для production нужна отдельная оценка CSRF, например Origin/Referer validation
или CSRF token. В classroom используем curl и не строим cross-origin frontend.
Разные origins потребуют согласовать credentials/CORS; SameSite относится к
site, а не к origin. SameSite=None требует Secure. Смотрите [security notes](docs/security-notes.md).

В SPA access живёт в памяти. После F5 он теряется, cookie остаётся:
`fetch('/auth/refresh', {method: 'POST', credentials: 'same-origin'})`
получает новый access и ротирует cookie. JS не считывает и не подставляет raw refresh.
Клиент пробует refresh один раз для подходящей 401 protected API, затем повторяет
request один раз. 401 от login означает неверный пароль, 401 от refresh — новый login.
Параллельные refresh следует объединять в одну клиентскую операцию.

## Endpoints

| Endpoint | Credential | Результат |
|---|---|---|
| GET /health | Public | 200 |
| POST /register | Email/password JSON | 201, duplicate 409 |
| POST /login | Email/password JSON | 200 + refresh cookie, invalid 401 |
| POST /auth/refresh | Refresh cookie | 200 + rotated cookie, invalid 401 |
| POST /auth/logout | Refresh cookie, без access | 204 + cleared cookie |
| GET /tasks | Bearer access | Свои задачи |
| GET /tasks/{id} | Bearer access | Своя Task, чужая 403 |
| POST /tasks | Bearer access | 201 + Location |
| PATCH /tasks/{id} | Bearer access | Прежняя partial-update semantics |
| DELETE /tasks/{id} | Bearer access | 204 |

Login и refresh JSON: `{"access_token":"...","token_type":"Bearer","expires_in":900}`.
Это **осознанное изменение login response** относительно занятия 14 (`token` стал
`access_token`), необходимое для явного различения credentials. Curl client и auth
fixtures обновлены. Task DTO `{id,title,done}`, routes, errors, PATCH и ownership прежние.
Refresh не возвращается в JSON. Auth responses имеют `Cache-Control: no-store`.

Logout отзывает текущую refresh session и очищает cookie с теми же Name/Path/SameSite/Secure.
В Go MaxAge=-1 сериализуется как HTTP Max-Age=0.
Повторный logout, неизвестная или отсутствующая cookie дают 204. При DB error — 500,
успешный logout не имитируем. Уже выданный stateless access работает до exp.

## Конфигурация

Go 1.22, PostgreSQL 16, psql, curl, Python 3 для demo scripts. Версии Go dependencies сохранены
из занятия 14, включая goose v3.22.1; новый toolchain не требуется.

| Env | Default / требование |
|---|---|
| DATABASE_URL | Обязательный DSN |
| TEST_DATABASE_URL | Отдельная БД, нужна только integration tests |
| JWT_SECRET | Случайный секрет, минимум 32 байта |
| ACCESS_TOKEN_TTL | 15m, целые секунды >= 1s |
| REFRESH_TOKEN_TTL | 168h, больше access TTL |
| REFRESH_COOKIE_NAME | refresh_token |
| COOKIE_SECURE | true, для local HTTP false |
| HTTP_ADDR | :8080 |
| LOG_LEVEL | INFO; DEBUG/WARN/ERROR/FATAL также поддерживаются |
| LOG_FORMAT | text или json |

`.env.example` содержит placeholders. Файл `.env` автоматически не загружается.
Вместо TOKEN_TTL занятия 14 теперь задаются два явных TTL. При restart сохраните JWT_SECRET.

## Запуск и migrations

Если PostgreSQL ещё не настроен, [пошаговый setup](docs/setup.md) создаёт две БД
с отдельным учебным owner. В уже рабочей БД занятия 14 сохраните DSN и JWT_SECRET.
Сначала остановите API14. В корне classroom15 выполните:

```bash
export DATABASE_URL='postgres://REPLACE_USER:REPLACE_PASSWORD@localhost:5432/REPLACE_DATABASE?sslmode=disable'
# Для существующей БД сохраните прежний JWT_SECRET; новый секрет нужен только новой установке.
export COOKIE_SECURE=false
make status
make up
make status
make run
```

`cmd/migrate/main.go` использует тот же goose Provider, что и занятие 14.
`001_init.sql` скопирован побайтно без изменений; `002_refresh_sessions.sql`
добавляет только refresh sessions. Оба файла содержат `Up` и `Down`,
`StatementBegin/StatementEnd`. История `goose_db_version` обеспечивает поведение:

- В новой пустой БД `make up` применяет 001 и 002.
- В базе занятия 14 с применённой 001 выполняется только 002.
- Повторный `make up` ничего не пересоздаёт; `make status` показывает версии.
- Один `make down` после версии 2 удаляет только refresh_sessions. **Повторный
  отдельный вызов `make down` удалит базовые users/tasks.** Данные удалённой
  таблицы не восстанавливаются при следующем Up.

Не редактируйте уже применённую 001 и не применяйте файлы через `psql -f`.
Старую базу, созданную вручную без goose history, сначала отдельно переводят
под управление миграциями; описанный автоматический upgrade рассчитан на базу
из текущего classroom14.

Полная схема после 002 уже находится в `structure.sql`. Обновление:
`make structure` вызывает прежний `scripts/dump-structure.sh` и реальный
`pg_dump --schema-only`, включая constraints, comments, indexes и goose table.
Для восстановления снимка требуется пустая целевая схема; это не миграция и
не замена goose history. Не загружайте snapshot поверх работающей базы.
При полном сбросе базы с повторным использованием user IDs смените JWT_SECRET;
при обычном restart сохраните прежний.

## Полный curl scenario

Обычный Task lifecycle: `make curl-demo` — прежние Alice/Bob/PATCH/ownership проверки,
теперь client читает access_token и в конце выполняет logout.

Для демонстрации expiration перезапустите тот же API с коротким access TTL:

```bash
export ACCESS_TOKEN_TTL=20s
export REFRESH_TOKEN_TTL=168h
export COOKIE_SECURE=false
make run
# In another terminal:
make session-demo
```

`session-demo.sh` создаёт Alice, логинится с `curl -c`, создаёт Task, ждёт exp
(максимум 30s), получает 401, выполняет `curl -b ... -c ... /auth/refresh`,
повторяет protected request, проверяет replay старого cookie → 401, logout → 204,
refresh после logout → 401 и работу уже выданного access до exp. Временные cookie
jars с правами только владельца удаляются автоматически, credentials не печатаются.
Пошаговые команды и проверка restart — [session-flow.md](docs/session-flow.md).

## Tests

```bash
make fmt-check
make ci
go test ./...
go vet ./...
go test -race ./...
# Explicit PostgreSQL tests; absence of TEST_DATABASE_URL is an error, not Skip:
export TEST_DATABASE_URL='postgres://REPLACE_USER:REPLACE_PASSWORD@localhost:5432/REPLACE_TEST_DATABASE?sslmode=disable'
make test-db
make test-db-race
```

UseCase/HTTP/JWT/middleware tests используют fake/memory dependencies.
DB tests создают изолированные schemas, применяют 001+002 и удаляют только свои schemas.
Они проверяют session CRUD, expiration/revoke, FK/cascade, UNIQUE, rotation,
rollback и concurrent refresh. Старые business/HTTP contract tests сохранены.

## Логирование

Сохранены `internal/logging`, его тесты и подход занятия 14. `LOG_LEVEL=DEBUG`
показывает операции без payload и credentials, INFO — успешные события, ERROR —
неожиданные сбои через тип ошибки. FATAL допустим только в main после возврата
run, чтобы cleanup успел выполниться. Не включайте raw DB errors или headers в
логи. Подробности — [logging.md](docs/logging.md).

## Материалы занятия

[Teacher guide](docs/teacher-guide.md) — маршрут на 85 минут с подготовленным boilerplate.
[Live coding](docs/live-coding-script.md), [continuity](docs/continuity.md),
[architecture](docs/architecture.md), [interview](docs/interview.md),
[bad examples](docs/bad-examples.md), [test strategy](docs/testing.md).
SQL 01–09 оставлены как reference занятия 14; новый SQL10 показывает state сессий без секретов.

## Границы обязательной реализации

Одна PostgreSQL DB, opaque refresh, sliding refresh lifetime, rotation, logout текущей session.
Нет Redis, OAuth/OIDC, MFA, IP/device binding, access denylist, logout-all или family replay detection.
Повтор старого refresh отвергается, но **не отзывает всех потомков**. Если credential
украден до rotation и атакующий использовал его первым, legitimate client получает
401. Автоматическое расследование/revocation всей family выходит за scope урока.
Управление сроком жизни/очисткой старых session rows — следующий production шаг.
