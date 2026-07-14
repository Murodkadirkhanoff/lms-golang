package data

import (
	"database/sql"

	"github.com/lib/pq"
)

type Models struct {
	Enrollments   EnrollmentModel
	Orders        OrderModel
	Certificates  CertificateModel
	Notifications NotificationModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Enrollments:   EnrollmentModel{DB: db},
		Orders:        OrderModel{DB: db},
		Certificates:  CertificateModel{DB: db},
		Notifications: NotificationModel{DB: db},
	}
}

func int64Array(ids []int64) any {
	return pq.Array(ids)
}
