// Package users is the composition root for the users bounded context. It wires
// the module's own layers and exposes its public contract to peers.
package users

import (
	"embed"
	"io/fs"

	coursescontract "github.com/chashma/lms/internal/modules/courses/contract"
	enrollmentcontract "github.com/chashma/lms/internal/modules/enrollment/contract"
	"github.com/chashma/lms/internal/modules/users/application"
	"github.com/chashma/lms/internal/modules/users/contract"
	"github.com/chashma/lms/internal/modules/users/infrastructure"
	"github.com/chashma/lms/internal/modules/users/transport"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Migrations returns this context's embedded migrations and version table.
func Migrations() (fs.FS, string, string) {
	return migrationsFS, "migrations", "schema_migrations_auth"
}

// MailConfig configures the password-reset mailer.
type MailConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	From        string
	FrontendURL string
}

// Module is the users bounded context.
type Module struct {
	service *application.Service
}

// New builds the users module. It depends only on framework ports (token
// issuer, mail config) — never on other bounded contexts.
func New(pool *pgxpool.Pool, issuer application.TokenIssuer, mail MailConfig) *Module {
	repo := infrastructure.NewUserRepository(pool)
	tokens := infrastructure.NewTokenRepository(pool)
	mailer := infrastructure.NewMailer(mail.Host, mail.Port, mail.Username, mail.Password, mail.From, mail.FrontendURL)
	return &Module{service: application.NewService(repo, tokens, issuer, mailer)}
}

// Directory exposes the users contract to peer modules.
func (m *Module) Directory() contract.UserDirectory { return m.service }

// RegisterRoutes mounts the HTTP endpoints. catalog/gate are peer contracts,
// injected here (after they are built) purely for the admin dashboard.
func (m *Module) RegisterRoutes(r chi.Router, catalog coursescontract.CourseCatalog, gate enrollmentcontract.EnrollmentGate) {
	transport.NewHandler(m.service, catalog, gate).Routes(r)
}
