// Package courses is the composition root for the courses bounded context.
package courses

import (
	"embed"
	"io/fs"

	enrollmentcontract "github.com/chashma/lms/internal/modules/enrollment/contract"
	userscontract "github.com/chashma/lms/internal/modules/users/contract"

	"github.com/chashma/lms/internal/modules/courses/application"
	"github.com/chashma/lms/internal/modules/courses/contract"
	"github.com/chashma/lms/internal/modules/courses/infrastructure"
	"github.com/chashma/lms/internal/modules/courses/transport"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrations returns this context's embedded migrations and version table.
func Migrations() (fs.FS, string, string) {
	return migrationsFS, "migrations", "schema_migrations_course"
}

// Module is the courses bounded context.
type Module struct {
	service *application.Service
}

// New builds the courses module. Its only cross-context dependency is the
// users directory (instructor names).
func New(pool *pgxpool.Pool, users userscontract.UserDirectory) *Module {
	svc := application.NewService(
		infrastructure.NewCourseRepository(pool),
		infrastructure.NewCategoryRepository(pool),
		infrastructure.NewQuizRepository(pool),
		infrastructure.NewReviewRepository(pool),
		infrastructure.NewQuestionRepository(pool),
		users,
	)
	return &Module{service: svc}
}

// Catalog exposes the courses contract to peer modules.
func (m *Module) Catalog() contract.CourseCatalog { return m.service }

// RegisterRoutes mounts the HTTP endpoints. gate (enrollment) is injected for
// review gating and the paywall.
func (m *Module) RegisterRoutes(r chi.Router, gate enrollmentcontract.EnrollmentGate) {
	transport.NewHandler(m.service, gate).Routes(r)
}
