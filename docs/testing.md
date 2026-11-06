# Test strategy

`go test ./...` не требует PostgreSQL. Старые Task business tests, Handler/router contract,
JWT algorithm/signature/expiry, bcrypt и middleware tests сохранены.
Только auth fixtures получили session dependencies и новое имя access_token.

Unit tests: генерация/длина/base64url без статистического randomness testing,
SHA-256 known vector, hash lookup вместо raw, configured expiry через now function,
invalid/expired/revoked/unknown, failure paths без выдачи credentials, logout.
Memory atomic rotation/copy isolation проверяется отдельно.
HTTP: login/refresh cookie attributes, no-store, новый refresh, replay401,
logout204/correct clearing, cookie-only /tasks401, valid access after logout200.
Expiration тесты не используют time.Sleep.

Explicit DB tests: `make test-db` и `make test-db-race`, DSN нужен обязательно.
Каждый test создаёт случайную schema в отдельной TEST_DATABASE_URL и выполняет обе migrations.
Session Create/Find/Revoke, expiration, unique, FK/cascade всей цепочки, failed INSERT
после revoke → полный rollback. Два simultaneous AuthUseCase.Refresh и отдельно
два direct repository Rotate дают ровно один success. Второй test не позволяет
ошибочно полагаться только на UseCase.Get без atomic repository operation.
SQL вызывается с request ctx, отменённый context проверяется на всех новых операциях.

Runtime QA: настоящий API + curl jars, expiry short access, refresh и retry,
old cookie replay, logout, 401/403/404/409, restart с прежним DB/JWT_SECRET:
User/Task/refresh session сохраняются. Новый login не требуется для активной refresh cookie.
Прежние Task contract responses проверяются после адаптации только login decoder.

Без изменений DB миграций unit/middleware tests остаются быстрыми.
CI: unit job без PostgreSQL + отдельный PostgreSQL16 job с explicit integration tags.

Goose проверяется тем же Provider, что и CLI. После 002 один Down сохраняет
users/tasks и удаляет только sessions; второй отдельный Down откатывает 001.
Добавлены tests, подтверждающие, что DEBUG/INFO/ERROR не раскрывают credentials.
