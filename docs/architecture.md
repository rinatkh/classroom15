# Архитектура и ответственность

Обычный request: Client → AuthMiddleware → Task Handler → Task UseCase → task.Repository → PostgresTaskRepository.
Authentication проверяет access JWT локально, ownership остаётся в Task UseCase.
Refresh cookie не даёт доступа к /tasks.

Auth endpoints: AuthHandler → user.UseCase (AuthUseCase) → user.Repository,
PasswordHasher, TokenIssuer, RefreshTokenGenerator и RefreshSessionRepository.
Naming user.UseCase сохранён с занятия 13: отдельный auth application package
ради нового имени не создаём. Cookie и JSON принадлежат Handler, SQL и transaction — repository.

RefreshSession содержит только hash и состояние. LoginResult включает raw refresh
только для передачи transport layer. Handler сериализует явный DTO, никогда не весь result.
`Rotate` — application contract atomic operation, без pgx.Tx снаружи.
Postgres реализация повторно проверяет active/expiration под row lock.
Memory реализация делает то же под mutex для быстрых tests.

В composition root по-прежнему создаются user/task PostgreSQL repositories,
hasher/JWTManager, UseCases, handlers/middleware/router. Добавляются session repository,
refresh generator и конфигурация TTL/cookie. TaskUseCase ничего об этом не знает.
