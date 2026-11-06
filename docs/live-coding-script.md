# Live coding: расширение готового Task API

Основной маршрут с минутами и слайдами — [teacher-guide.md](teacher-guide.md).
Полная устная речь находится в заметках презентации. Порядок кода:

1. В отдельном classroom14 выполнить make run, make curl-demo, остановить API.
   Bob → 403. Дальше работать в единственном корне classroom15.
2. Открыть internal/task/usecase.go и internal/httpapi/middleware/auth.go:
   они остаются знакомыми. Разобрать internal/user/session.go, поля и port.
3. 001_init.sql совпадает с занятием 14. Дополнить схему только через
   002_refresh_sessions.sql в формате goose Up/Down. Выполнить make status,
   make up, make status. База 14 выполняет только 002.
4. Make structure: найти users/tasks/refresh_sessions, FK/UNIQUE/index/comments.
   Down демонстрировать на disposable DB: один вызов откатывает 002, второй
   отдельный вызов удалит 001 и данные. Восстановить make up.
5. internal/auth/refresh.go: Generate и Hash; go test ./internal/auth -v.
   Сопоставить SHA-256 random credential и прежний bcrypt password.
6. internal/user/usecase.go: Login возвращает результат только после Create.
   Raw не входит в persistence model.
7. internal/httpapi/handler/auth.go, cookie.go: explicit DTO, HttpOnly/Secure/
   SameSite/Path и no-store. Не сериализовать целиком LoginResult.
8. UseCase.Refresh: предварительные проверки. Postgres.Rotate: transaction,
   row lock, recheck после ожидания, revoke/insert/link/commit и rollback.
9. Make test-db, make test-db-race: failed INSERT сохраняет old; два прямых
   Rotate создают только одного успешного потомка.
10. Logout: idempotent Revoke, matching cookie deletion и 204. Access до exp
    может действовать; это подтверждено HTTP test.
11. Cmd/api/main.go: добавлены auth dependencies и config; Task wiring и logger
    сохранены. Make ci.
12. ACCESS_TOKEN_TTL=20s COOKIE_SECURE=false make run; во втором терминале
    make curl-demo, make session-demo. Отдельно проверить active jar после restart
    с прежними БД/JWT_SECRET.
13. Снова открыть Task UseCase: refresh не проник в business logic. Homework
    остаётся отдельным модулем со своими TODO и acceptance tests.

Точки возврата к презентации: 3, 15, 17, 19, 21, 23, 26, 28, 30, 31, 35,
39, 40, 43, 44. Для каждого перехода конкретные команды и проверяемый результат
также записаны в Speaker Notes соответствующего слайда.
