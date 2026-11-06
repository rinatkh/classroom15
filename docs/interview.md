# Interview с ответами

1. **Почему access короткий?** При утечке ограничивается окно использования.
   Key point: TTL уменьшает impact, но не предотвращает кражу. Ошибка: «JWT сам безопасен бесконечно».
2. **Почему access не в DB, а refresh session в DB?** JWT проверяем подписью/exp;
   долгую сессию нужно отзывать и ротировать. Key point: stateless access, stateful refresh.
3. **Зачем хранить hash?** Утечка DB не должна сразу дать готовый refresh credential.
   Key point: клиент имеет raw, сервер ищет digest. Ошибка: возвращать hash как token.
4. **Почему bcrypt и SHA-256 используются по-разному?** Password слабый, нужен дорогой salted hash.
   Refresh случайный 256-bit secret, deterministic SHA-256 подходит для lookup.
   Key point: выбор определяется entropy и задачей, не скоростью реализации.
5. **HttpOnly защищает от CSRF?** Нет, он запрещает JS читать cookie. Браузер всё ещё прикладывает её автоматически.
   Key point: SameSite и CSRF policy — отдельные меры.
6. **Access ещё жив после logout?** Да, в текущей stateless модели до exp.
   Key point: отзываем refresh session, не JWT. Ошибка: обещать мгновенный logout всех requests.
7. **Зачем transaction и row lock?** Проверка и замена session должны быть атомарными при двух requests.
   Key point: ровно один winner, второй видит revoked. Ошибка: защищаться Go mutex между разными API instances.
8. **Зачем повторная проверка в repository после UseCase.Get?** Между чтением и записью другой request мог изменить row.
   Key point: application объясняет правило, persistence гарантирует atomicity.
9. **Что делает F5 в SPA?** Access в памяти исчезает, startup refresh по HttpOnly cookie выдаёт новый.
   Key point: JS не считывает raw refresh.
10. **Почему не каждый 401 ведёт к refresh?** Login401 — credentials, refresh401 — нужен login;
    подходящая protected401 допускает один refresh/retry. Key point: ограниченный client flow.
