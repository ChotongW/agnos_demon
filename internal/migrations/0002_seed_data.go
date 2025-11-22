package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var migration0002SeedData = &Migration{
	Number: 2,
	Name:   "Seed initial data",
	Forwards: func(db *pgxpool.Pool, logger *logrus.Logger) error {
		ctx := context.Background()

		sql := `
			-- Super Admin (Password: password)
			INSERT INTO staff (id, username, password_hash, hospital) VALUES 
			(uuid_generate_v4(), 'admin', '$2a$12$VsfCQivbKsbMdc8i9jMTTO2ekdf7FBjIH9r8X1SH4UG6GFZVNsnsK', 'hn-001');

			-- Staff for Hospital B (Password: password)
			INSERT INTO staff (id, username, password_hash, hospital) VALUES 
			(uuid_generate_v4(), 'staff_b', '$2a$12$VsfCQivbKsbMdc8i9jMTTO2ekdf7FBjIH9r8X1SH4UG6GFZVNsnsK', 'hn-002');

			-- Patients for Hospital A
			INSERT INTO patients (id, patient_hn, first_name_th, last_name_th, first_name_en, last_name_en, date_of_birth, gender, national_id, passport_id) VALUES
			(uuid_generate_v4(), 'hn-001', 'จอห์น', 'โด', 'John', 'Doe', '1980-01-01', 'M', '9855629944793', 'AB123456'),
			(uuid_generate_v4(), 'hn-001', 'เจน', 'สมิธ', 'Jane', 'Smith', '1990-05-15', 'F', '9220297763701', 'CD654321');

			-- Patients for Hospital B
			INSERT INTO patients (id, patient_hn, first_name_th, last_name_th, first_name_en, last_name_en, date_of_birth, gender, national_id, passport_id) VALUES
			(uuid_generate_v4(), 'hn-002', 'อลิซ', 'จอห์นสัน', 'Alice', 'Johnson', '1985-03-20', 'F', '3753395384991', 'EF567890'),
			(uuid_generate_v4(), 'hn-002', 'บ็อบ', 'บราวน์', 'Bob', 'Brown', '1975-11-10', 'M', '5258439754182', 'GH123456');
		`

		_, err := db.Exec(ctx, sql)
		if err != nil {
			return err
		}

		logger.Info("Seed data inserted successfully")
		return nil
	},
}

func init() {
	Migrations = append(Migrations, migration0002SeedData)
}
