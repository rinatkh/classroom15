# Ошибки, которые разбираем на занятии

| Ошибка | Последствие | Правильное место/решение |
|---|---|---|
| Refresh JWT принимается как access | Размывается назначение credentials | Opaque refresh только на auth endpoints |
| Raw refresh в PostgreSQL или JSON | Утечка credentials | SHA-256 в DB, HttpOnly cookie клиенту |
| math/rand для refresh | Предсказуемые значения | crypto/rand |
| bcrypt для прямого hash lookup | Salt даёт разные digest | bcrypt для password; SHA-256 для random refresh |
| Только SELECT active, потом INSERT | Две дочерние session при concurrency | Transaction + FOR UPDATE + повторная проверка |
| Revoke old вне transaction | Сбой INSERT ломает сессию | Rotate atomic repository operation |
| now() до долгого ожидания lock | Expiration проверяется на устаревшее время | clock_timestamp после row lock |
| http.Cookie в UseCase | Application зависит от transport | Cookie helper в Handler |
| pgx.Tx в UseCase | SQL протекает во внутренний слой | Rotate contract без pgx |
| Middleware автоматически refresh | Скрытый DB request на каждый API call | Явный POST /auth/refresh |
| Refresh cookie открывает /tasks | Долгий credential используется повсюду | Bearer access остаётся обязательным |
| Любой 401 вызывает бесконечный retry | Login/refresh loop | Один осмысленный refresh и один retry |
| Logout обещает мгновенно убить access | Неверная модель stateless JWT | Он живёт до exp |
| HttpOnly объясняется как CSRF защита | Ложная гарантия | Разделить чтение cookie и автоматическую отправку |
