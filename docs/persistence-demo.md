# Persistence после занятия 15

Основной маршрут находится в [session-flow.md](session-flow.md).
При restart с тем же DSN/JWT_SECRET сохраняются users, tasks и refresh_sessions.
Обычный restart не равен db-down: migrations откатывать не нужно.
