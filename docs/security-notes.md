# Security notes: границы учебного решения

Password: bcrypt с прежним cost. Email lower-case по прежнему UseCase; DB UNIQUE — final protection.
Access: прежний HS256 JWT с обязательным exp, ограничением algorithm, положительным user_id.
Секрет минимум 32 байта, хранится вне git. Access проверяется без refresh_sessions.
Refresh: crypto/rand 32 bytes = 256 bits; unpadded base64url имеет длину 43 characters.
В БД только SHA-256 hex, ни password, ни access JWT, ни raw refresh там не сохраняются.

HttpOnly затрудняет кражу cookie через JS, но XSS всё ещё может выполнять requests
от имени пользователя и видеть access в памяти. HttpOnly не защищает от CSRF.
SameSite=Lax — дополнительное ограничение cross-site cookie sending, а не универсальная CSRF защита.
Same-origin demo не включает CORS. Для реального frontend отдельно проверяют Origin/Referer
или CSRF token, CORS allowlist с credentials, cookie SameSite/Secure и HTTPS.
Origin и site различаются: разные порты — разные origins, но могут быть одним site.
Cookie Path=/auth уменьшает её доставку, а не создаёт security boundary.

COOKIE_SECURE=true по умолчанию. false только для local HTTP.
Cookie host-only без Domain. При logout сохраняются Name/Path/Secure/SameSite/HttpOnly,
MaxAge=-1 и Expires в прошлом. __Host- prefix несовместим с выбранным Path=/auth,
поэтому config запрещает его. Auth responses no-store.

Logout idempotent и отзывает только предъявленную refresh session. Stateless access
после logout работает до exp. Denylist в scope не входит. При старом refresh после
rotation — 401, потомки автоматически не отзываются. Успешный refresh продлевает
refresh lifetime на configured TTL; абсолютный предел всей цепочки не реализован.

Сетевой сбой после COMMIT, но до доставки нового cookie может потребовать новый login:
старый credential уже revoked. Не ретраим refresh бесконечно. Клиент сериализует refresh,
при 401 от refresh переходит к login. Это также предотвращает гонку двух вкладок в demo.
Реальные multi-tab coordination, family reuse detection и очистка expired rows — дальнейшая работа.

Не логировать access/refresh, Authorization/Cookie/Set-Cookie, password или password_hash.
Demo scripts выводят только название проверки и status. Не запускайте их с `bash -x`.
Cookie jars — секреты, права 0600, удалить после ручного demo.
SQL10 показывает state без token_hash. DB errors наружу 500 с нейтральным текстом.

Первоисточники:

- [Go crypto/rand](https://pkg.go.dev/crypto/rand@go1.22.12), [crypto/sha256](https://pkg.go.dev/crypto/sha256).
- [MDN Set-Cookie](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Set-Cookie).
- [OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).
- [OWASP CSRF Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).
- [PostgreSQL16 row locks](https://www.postgresql.org/docs/16/explicit-locking.html), [time functions](https://www.postgresql.org/docs/16/functions-datetime.html).

Это собственный учебный session flow, а не реализация OAuth2/OIDC protocol.
