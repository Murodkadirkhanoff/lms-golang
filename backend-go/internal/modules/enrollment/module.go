// Package enrollment is the composition root for the enrollment bounded context.
package enrollment

import (
	"embed"
	"io/fs"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	userscontract "github.com/chashma/lms/internal/modules/users/contract"

	"github.com/chashma/lms/internal/modules/enrollment/application"
	"github.com/chashma/lms/internal/modules/enrollment/contract"
	"github.com/chashma/lms/internal/modules/enrollment/infrastructure"
	"github.com/chashma/lms/internal/modules/enrollment/transport"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrations returns this context's embedded migrations and version table.
func Migrations() (fs.FS, string, string) {
	return migrationsFS, "migrations", "schema_migrations_enrollment"
}

// Module is the enrollment bounded context.
type Module struct {
	service *application.Service
}

// New builds the enrollment module. It depends on the courses catalog and
// users directory through their contracts.
func New(pool *pgxpool.Pool, catalog coursescontract.CourseCatalog, users userscontract.UserDirectory) *Module {
	svc := application.NewService(
		infrastructure.NewEnrollmentRepository(pool),
		infrastructure.NewOrderRepository(pool),
		infrastructure.NewCertificateRepository(pool),
		infrastructure.NewNotificationRepository(pool),
		infrastructure.NewPDFRenderer(),
		catalog,
		users,
	)
	return &Module{service: svc}
}

// Gate exposes the enrollment contract to peer modules.
func (m *Module) Gate() contract.EnrollmentGate { return m.service }

// RegisterRoutes mounts the HTTP endpoints.
func (m *Module) RegisterRoutes(r chi.Router) {
	transport.NewHandler(m.service).Routes(r)
}
