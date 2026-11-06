# Continuity: финал 14 → финал 15

Исходная точка — отдельный завершённый classroom14. Classroom15 содержит один
модуль без вложенных snapshots и без homework. Занятие 14 не изменялось.

## Сохранено

Task model, task.Repository, task.UseCase и все business/ownership tests;
Task DTO/Mapper/Handler и PATCH contract; Postgres User/Task repositories;
User model/Repository, нормализация email и bcrypt; JWTManager/TokenVerifier;
AuthMiddleware и auth context; прежние public /health, /register, /login и пять
protected Task routes. Сохранены logging infrastructure и её тесты, goose runner,
SQL labs 01–09 и механизм schema dump.

## Расширено по необходимости

| Файлы | Причина |
| --- | --- |
| internal/user/session.go, model.go, errors.go, usecase.go | RefreshSession, LoginResult, новые ошибки, Login/Refresh/Logout |
| internal/auth/refresh.go | crypto/rand + SHA-256 для opaque credential |
| internal/repository/postgres/session.go | Реализация нового port; атомарный Rotate |
| internal/repository/memory/session.go | Быстрые tests без PostgreSQL |
| internal/httpapi/dto/auth.go | Явный access response вместо прежнего token |
| internal/httpapi/handler/auth.go, cookie.go | Cookie на HTTP-границе, refresh/logout, no-store |
| internal/httpapi/router.go | Два новых public POST endpoint |
| cmd/api/main.go, .env.example, Makefile | Новые concrete dependencies и два TTL; прежний logging сохранён |
| migrations/002_refresh_sessions.sql | Только новая таблица, индексы и комментарии |
| structure.sql | Реальный schema-only dump после двух миграций |
| auth fixtures, новые session tests | Новый constructor/response; прежние Task assertions сохранены |
| docs и scripts | Goose history, cookie jar, restart и демонстрация rotation |

`migrations/001_init.sql` идентичен файлу занятия 14 побайтно. Унаследованные
пакеты отличаются только module path `classroom14` → `classroom15` там, где это
необходимо для самостоятельного Go-модуля. Task Handler/UseCase и storage
не переписывались ради refresh. В Task unit fixture изменён только constructor
AuthHandler; assertions сохранены. У migration test проверяются обе границы: первый Down удаляет только sessions,
второй отдельный Down — базовую 001; затем Up восстанавливает схему.

Login HTTP response осознанно изменился с `{"token":"..."}` на
`{"access_token":"...","token_type":"Bearer","expires_in":900}` и Set-Cookie.
Task DTO и ownership не изменились. Обычная проверка access не читает sessions;
сами Task операции по-прежнему обращаются в PostgreSQL через UseCase.

Goose history применяет только 002 на существующей базе 14; при новой базе —
001 и 002. Один Down возвращает схему 14 с сохранёнными users/tasks.
PostgreSQL, HTTP cookies и логирование не добавлены во внутренние UseCases.

## Пофайловый diff относительно финала 14

Ниже перечислены все файлы, чьё содержимое отличается после нормализации имени модуля и путей classroom14 → classroom15. Остальные унаследованные файлы сохранены.

| Файл | Изменение | Причина |
| --- | --- | --- |
| `.env.example` | изменён | Два TTL, имя cookie, Secure; прежние DB/JWT/logging параметры сохранены. |
| `.gitignore` | изменён | Не хранить локальную конфигурацию и бинарники; убрать устаревшие дубли. |
| `Makefile` | изменён | Сохранить команды 14; добавить session demo и сборку текущего модуля. |
| `README.md` | изменён | Единый запуск, goose history, новый auth lifecycle и прежний Task API. |
| `cmd/api/main.go` | изменён | Единственное место подключения новых concrete session dependencies, TTL и cookie config. |
| `docs/architecture.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/bad-examples.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/continuity.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/interview.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/live-coding-script.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/persistence-demo.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/security-notes.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/session-flow.md` | добавлен | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/setup.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/sql-cheatsheet.md` | изменён | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/teacher-guide.md` | добавлен | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `docs/testing.md` | добавлен | Согласовать объяснение, команды и маршрут практики с текущим session lifecycle. |
| `internal/auth/refresh.go` | добавлен | Opaque token из crypto/rand; SHA-256 digest для хранения. |
| `internal/auth/refresh_test.go` | добавлен | Проверить генерацию, формат, размер, уникальность и hash. |
| `internal/httpapi/dto/auth.go` | изменён | Access DTO не сериализует raw refresh и внутреннюю session. |
| `internal/httpapi/handler/auth.go` | изменён | HTTP refresh/logout, DTO, cookie и нейтральный error mapping; безопасные логи. |
| `internal/httpapi/handler/auth_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/httpapi/handler/cookie.go` | добавлен | HTTP-only cookie configuration, запись и удаление с тем же Path. |
| `internal/httpapi/handler/logging_test.go` | изменён | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/httpapi/handler/task_test.go` | изменён | Только адаптация NewAuthHandler fixture; прежние Task assertions сохранены. |
| `internal/httpapi/router.go` | изменён | Добавить два public POST auth endpoint; прежние Task routes неизменны. |
| `internal/httpapi/router_test.go` | изменён | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/httpapi/session_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/repository/memory/session.go` | добавлен | Новый memory adapter для быстрых изолированных unit tests. |
| `internal/repository/memory/session_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/repository/postgres/application_integration_test.go` | изменён | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/repository/postgres/repository_integration_test.go` | изменён | Сохранить прежние repository tests и idempotent Up; проверить двухступенчатый Down. |
| `internal/repository/postgres/session.go` | добавлен | Новый PostgreSQL adapter; транзакционная rotation и безопасные operation logs. |
| `internal/repository/postgres/session_integration_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/user/errors.go` | изменён | Application errors для refresh, срока действия и отзыва. |
| `internal/user/model.go` | изменён | LoginResult теперь содержит access metadata и внутренние refresh данные. |
| `internal/user/session.go` | добавлен | Новые session model, Repository port, Generator port и TTL config. |
| `internal/user/session_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `internal/user/usecase.go` | изменён | Расширить Login; добавить Refresh/Logout, не меняя Task UseCase. |
| `internal/user/usecase_test.go` | изменён | Сохранить Register/Login тесты и адаптировать новые dependencies/result. |
| `migrations/002_refresh_sessions.sql` | добавлен | Единственная новая goose миграция; Up/Down с аннотациями, FK/UNIQUE/index/comments. |
| `scripts/curl-demo.sh` | изменён | Воспроизводимая демонстрация access/cookie jar, rotation и logout без утечки credentials. |
| `scripts/session-demo.sh` | добавлен | Воспроизводимая демонстрация access/cookie jar, rotation и logout без утечки credentials. |
| `sql/10_refresh_sessions.sql` | добавлен | SQL experiment для refresh sessions; не меняет production Task API. |
| `structure.sql` | изменён | Полный schema-only pg_dump PostgreSQL 16 после 001+002, без пользовательских данных. |
| `test/integration/contract_test.go` | добавлен | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
| `test/integration/fixture_test.go` | изменён | Проверки refresh/cookie/session/rotation и сохранения прежнего контракта. |
