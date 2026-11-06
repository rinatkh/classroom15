# Занятие 15: маршрут на 90 минут

Исходная точка — отдельный classroom14. Здесь один classroom15 без snapshots
и без homework. Полная речь находится в Speaker Notes всех 44 слайдов и в
приложенном lesson15_teacher_speech.md. Этот файл задаёт маршрут работы.

## Подготовка

Открыть classroom14 и classroom15 отдельно в GoLand. Подготовить PostgreSQL,
прежние DATABASE_URL/JWT_SECRET и отдельную TEST_DATABASE_URL. Заранее выполнить
make ci, make test-db и make test-db-race в classroom15. API14 и API15 не
должны одновременно слушать один порт. Credentials не печатать; bash -x не включать.

| Минуты | Слайды | Файлы и команды | Результат и возврат |
| --- | --- | --- | --- |
| 0–7 | 1–2 | В отдельном classroom14: make run, make curl-demo | Task CRUD, Bob → 403; остановить API14 → 3 |
| 7–19 | 3–13 | Готовые схемы client/access/cookie | Назначение credentials, границы HttpOnly |
| 19–28 | 14–16 | internal/user/session.go, migrations/002_refresh_sessions.sql; make status, make up, make status, make structure | 001 прежняя, 002 новая → 17 |
| 28–41 | 17–22 | internal/auth/refresh.go; user/usecase.go Login; handler/auth.go и cookie.go | Hash-only save и credentials после Create → 23 |
| 41–57 | 23–29 | user/usecase.go Refresh; postgres/session.go Rotate; make test-db | Recheck под lock, rollback, прежний Task → 30 |
| 57–65 | 30–36 | Logout, cookie clearing, cmd/api/main.go | 204, access до exp, DI и continuity → 37 |
| 65–74 | 37–38 | session_integration_test.go; make test-db-race | Один concurrent winner, настоящая БД → 39 |
| 74–85 | 39–40 | ACCESS_TOKEN_TTL=20s COOKIE_SECURE=false make run; второй терминал make curl-demo, make session-demo | Expired 401, rotation, replay 401, logout |
| 85–90 | 41–44 | В отдельном homework_go_15: make starter-check, make test-acceptance | Starter собирается, TODO acceptance красный |

Не набирать весь проект. Live coding — новая migration, Generate/Hash,
расширение Login и ключевые строки cookie/Rotate. Остальной подготовленный код
читать по invariants и проверять tests. Полный CI можно выполнить до занятия.

## Контрольные вопросы

- Слайд 4: что принимает GET /tasks? Только Bearer access.
- Слайд 9: Login на /login может установить Path=/auth? Да, Path управляет
  последующей доставкой cookie; Tasks не авторизуются refresh-cookie.
- Слайд 14: RevokedAt=nil гарантирует active? Нет, нужен ExpiresAt в будущем.
- Слайд 16: второй отдельный Down? Удалит базовые users/tasks и их данные.
- Слайд 21: failed Create позволяет отправить cookie? Нет, credentials не выдаются.
- Слайд 28: B после lock доверяет актуальной строке, а не старому application Get.
- Слайд 31: access может действовать после logout? Да, до exp при stateless проверке.

## Точечные команды

```bash
go test ./internal/auth -v
go test ./internal/user -run 'TestLogin|TestRefresh' -v
go test ./internal/httpapi -run TestSessionHTTPFlow -v
go test -tags=integration ./internal/repository/postgres -run TestIntegrationRotationRollbackAndIdentity -v
go test -tags=integration ./internal/repository/postgres -run TestIntegrationConcurrentRepositoryRotateOnlyOneSuccess -v
```

Tagged tests требуют отдельной TEST_DATABASE_URL. Fake не доказывает row lock,
а Go race detector не является проверкой SQL atomicity.

## Restart demonstration

Сделать отдельный Login, сохранить active cookie jar. Остановить API и запустить
с прежними DATABASE_URL/JWT_SECRET. Refresh из этого jar должен вернуть 200;
сохранённая Task читается новым access. Не брать jar, отозванный session-demo.
Состояние без credentials: psql "$DATABASE_URL" -X -f sql/10_refresh_sessions.sql.

Вернуться к Task UseCase: в нём нет refresh. Новые concrete dependencies
собраны в main; logger, goose и dump сохранены. Полная речь и точные переходы
по слайдам — в Speaker Notes.
