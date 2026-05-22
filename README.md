# javar
Сайт з беларусізатарамі

## Admin auth

Адмінка выкарыстоўвае пароль і HttpOnly cookie-сесію. У `.env` патрэбныя:

```env
ADMIN_PASSWORD_HASH=<bcrypt hash>
SESSION_SECRET=<long random string>
```

Хэш пароля можна згенераваць так:

```sh
go run ./cmd/hash-password "your-admin-password"
```
