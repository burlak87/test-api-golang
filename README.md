Осталось:
<!-- Реализовать защиту от XSS  -->
<!-- Реализовать rate limiting для эндпоинтов аутентификации (максимум 5 попыток в минуту) -->

Tech Information:
- Vanilla Go
- PostgreSQL
- JWT
- bcrypt или альтернатива

Model:
Дипломная работа
Студент
Связь один к одному

routing: 
- POST api/auth/register
- POST api/auth/login
- POST api/auth/refresh
<!-- - POST api/resources -->
- GET api/resources?page=1&limit=10
<!-- - GET api/resources/{id}
- PUT api/resources/{id}
- DELETE api/resources/{id} -->

Требования: 
Эндпоинты требуют валидный JWT Token, соответственно кроме аутентификации
Защита от XSS
Безопасность аутентификации, а именно не более 5 попыток за 2 минуты, пароль от 8ми символов.

go get -t github.com/mitchellh/mapstructure/
go get github.com/gorilla/mux
<!-- go get github.com/julienschmidt/httprouter -->
go get github.com/sirupsen/logrus@v1.9.3
go get -u github.com/ilyakaznacheev/cleanenv
go get github.com/jackc/pgconn && go get github.com/jackc/pgx && go get github.com/jackc/pgx/v4
<!-- go get github.com/jackc/pgx@none -->
go get github.com/jackc/pgx/v4/pgxpool@v4.18.3

go clean -modcache

<!-- supabase -->
go get github.com/jackc/pgx/v5