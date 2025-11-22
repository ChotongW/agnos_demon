package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

var migration0001InitialSchema = &Migration{
	Number: 1,
	Name:   "Create initial schema",
	Forwards: func(db *pgxpool.Pool, logger *logrus.Logger) error {
		ctx := context.Background()

		sql := `
			CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

			CREATE TABLE staff (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				username VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				hospital VARCHAR(255) NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);

			CREATE TYPE gender_enum AS ENUM ('M', 'F');

			CREATE TABLE patients (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
				patient_hn VARCHAR(255) NOT NULL,
				first_name_th VARCHAR(255),
				middle_name_th VARCHAR(255),
				last_name_th VARCHAR(255),
				first_name_en VARCHAR(255),
				middle_name_en VARCHAR(255),
				last_name_en VARCHAR(255),
				date_of_birth DATE,
				national_id VARCHAR(13) UNIQUE,
				passport_id VARCHAR(20) UNIQUE,
				phone_number VARCHAR(20),
				email VARCHAR(255) UNIQUE,
				gender gender_enum,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
			);

			CREATE INDEX idx_patients_patient_hn ON patients(patient_hn);
			CREATE INDEX idx_staff_hospital ON staff(hospital);
			CREATE INDEX idx_staff_username ON staff(username);
			CREATE INDEX idx_patients_national_id ON patients(national_id);
			CREATE INDEX idx_patients_passport_id ON patients(passport_id);
		`

		_, err := db.Exec(ctx, sql)
		if err != nil {
			return err
		}

		logger.Info("Initial schema created successfully")
		return nil
	},
}

func init() {
	Migrations = append(Migrations, migration0001InitialSchema)
}
