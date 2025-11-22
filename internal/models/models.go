package models

import (
	"time"

	"github.com/google/uuid"
)

type Staff struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Hospital     string    `json:"hospital"`
	CreatedAt    time.Time `json:"created_at"`
}

type Date struct {
	time.Time
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + d.Time.Format("2006-01-02") + `"`), nil
}

func (d *Date) Scan(value interface{}) error {
	if value == nil {
		d.Time = time.Time{}
		return nil
	}
	if t, ok := value.(time.Time); ok {
		d.Time = t
		return nil
	}
	return nil
}

type Patient struct {
	ID           uuid.UUID  `json:"id"`
	PatientHN    string     `json:"patient_hn"` // Hospital Number
	FirstNameTH  string     `json:"first_name_th"`
	MiddleNameTH string     `json:"middle_name_th"`
	LastNameTH   string     `json:"last_name_th"`
	FirstNameEN  string     `json:"first_name_en"`
	MiddleNameEN string     `json:"middle_name_en"`
	LastNameEN   string     `json:"last_name_en"`
	DateOfBirth  *Date      `json:"date_of_birth"`
	Gender       string     `json:"gender"`
	NationalID   string     `json:"national_id"`
	PassportID   string     `json:"passport_id"`
	PhoneNumber  string     `json:"phone_number"`
	Email        string     `json:"email"`
	CreatedAt    *time.Time `json:"-"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

type CreateStaffRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Hospital string `json:"hospital" binding:"required"`
}

type SearchPatientResponse struct {
	Patient []*Patient `json:"patients"`
}
