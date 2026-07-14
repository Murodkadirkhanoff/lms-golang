module lms.chashma.uz/course-service

go 1.26.2

require (
	github.com/go-chi/chi/v5 v5.3.1
	github.com/lib/pq v1.12.3
	lms.chashma.uz/pkg v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang-migrate/migrate/v4 v4.19.1 // indirect
)

replace lms.chashma.uz/pkg => ../pkg
