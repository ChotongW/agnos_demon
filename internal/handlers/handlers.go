package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"agnos_demo/internal/database"
	"agnos_demo/internal/middleware"
	"agnos_demo/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Handlers struct {
	db     database.DB
	logger *slog.Logger
}

func NewHandlers(db database.DB, logger *slog.Logger) *Handlers {
	return &Handlers{
		db:     db,
		logger: logger,
	}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "OK"})
}

func (h *Handlers) CreateStaff(c *gin.Context) {
	var input models.CreateStaffRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("Invalid staff creation request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Creating staff", "username", input.Username, "hospital", input.Hospital)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	ctx := context.Background()
	var staffID uuid.UUID

	query := `
		INSERT INTO staff (username, password_hash, hospital)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err = h.db.QueryRow(ctx, query, input.Username, string(hashedPassword), input.Hospital).Scan(&staffID)
	if err != nil {
		h.logger.Error("Failed to create staff in database", "error", err, "username", input.Username)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create staff"})
		return
	}

	h.logger.Info("Staff created successfully", "staff_id", staffID, "username", input.Username, "hospital", input.Hospital)
	c.JSON(http.StatusCreated, gin.H{"message": "Staff created successfully", "id": staffID})
}

func (h *Handlers) LoginStaff(c *gin.Context) {
	var input models.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Warn("Invalid login request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Debug("Login attempt", "username", input.Username, "hospital", input.Hospital)

	ctx := context.Background()
	var staff models.Staff

	query := `
		SELECT id, username, password_hash, hospital
		FROM staff
		WHERE username = $1 AND hospital = $2
	`

	err := h.db.QueryRow(ctx, query, input.Username, input.Hospital).Scan(
		&staff.ID,
		&staff.Username,
		&staff.PasswordHash,
		&staff.Hospital,
	)
	if err != nil {
		h.logger.Warn("Login failed - user not found", "username", input.Username, "hospital", input.Hospital, "error", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(staff.PasswordHash), []byte(input.Password)); err != nil {
		h.logger.Warn("Login failed - invalid password", "username", input.Username, "hospital", input.Hospital)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := middleware.GenerateToken(staff.ID.String(), staff.Hospital)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err, "staff_id", staff.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	h.logger.Info("Login successful", "staff_id", staff.ID, "username", input.Username, "hospital", input.Hospital)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handlers) SearchPatient(c *gin.Context) {
	hospital, exists := c.Get("hospital")
	if !exists {
		h.logger.Warn("Unauthorized patient search attempt - no hospital in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	h.logger.Debug("Patient search request", "hospital", hospital, "query_params", c.Request.URL.RawQuery)

	ctx := context.Background()

	var conditions []string
	var args []interface{}
	argIndex := 1

	hospitalFilter := hospital.(string)
	conditions = append(conditions, fmt.Sprintf("patient_hn LIKE $%d", argIndex))
	args = append(args, hospitalFilter+"%")
	argIndex++

	if patientHN := c.Query("patient_hn"); patientHN != "" {
		conditions = append(conditions, fmt.Sprintf("patient_hn = $%d", argIndex))
		args = append(args, patientHN)
		argIndex++
	}
	if nationalID := c.Query("national_id"); nationalID != "" {
		conditions = append(conditions, fmt.Sprintf("national_id = $%d", argIndex))
		args = append(args, nationalID)
		argIndex++
	}
	if passportID := c.Query("passport_id"); passportID != "" {
		conditions = append(conditions, fmt.Sprintf("passport_id = $%d", argIndex))
		args = append(args, passportID)
		argIndex++
	}
	if firstName := c.Query("first_name"); firstName != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(first_name_en) LIKE LOWER($%d) OR LOWER(first_name_th) LIKE LOWER($%d))", argIndex, argIndex))
		args = append(args, "%"+firstName+"%")
		argIndex++
	}
	if middleName := c.Query("middle_name"); middleName != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(middle_name_en) LIKE LOWER($%d) OR LOWER(middle_name_th) LIKE LOWER($%d))", argIndex, argIndex))
		args = append(args, "%"+middleName+"%")
		argIndex++
	}
	if lastName := c.Query("last_name"); lastName != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(last_name_en) LIKE LOWER($%d) OR LOWER(last_name_th) LIKE LOWER($%d))", argIndex, argIndex))
		args = append(args, "%"+lastName+"%")
		argIndex++
	}
	if dob := c.Query("date_of_birth"); dob != "" {
		conditions = append(conditions, fmt.Sprintf("date_of_birth = $%d", argIndex))
		args = append(args, dob)
		argIndex++
	}

	query := fmt.Sprintf(
		`SELECT id, patient_hn, first_name_th, middle_name_th, last_name_th, first_name_en, middle_name_en, last_name_en, 
		 date_of_birth, gender, national_id, passport_id, phone_number, email 
		 FROM patients WHERE %s`,
		strings.Join(conditions, " AND "),
	)

	h.logger.Debug("Executing patient search query", "hospital", hospital, "conditions_count", len(conditions))

	rows, err := h.db.Query(ctx, query, args...)
	if err != nil {
		h.logger.Error("Failed to execute patient search query", "error", err, "hospital", hospital)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch patients"})
		return
	}
	defer rows.Close()

	var patients []*models.Patient
	for rows.Next() {
		var p models.Patient
		var (
			firstNameTH, middleNameTH, lastNameTH              *string
			firstNameEN, middleNameEN, lastNameEN              *string
			gender, nationalID, passportID, phoneNumber, email *string
			dateOfBirth                                        *models.Date
		)

		err := rows.Scan(
			&p.ID, &p.PatientHN,
			&firstNameTH, &middleNameTH, &lastNameTH,
			&firstNameEN, &middleNameEN, &lastNameEN,
			&dateOfBirth, &gender,
			&nationalID, &passportID, &phoneNumber, &email,
		)
		if err != nil {
			h.logger.Error("Failed to scan patient row", "error", err)
			continue
		}

		if firstNameTH != nil {
			p.FirstNameTH = *firstNameTH
		}
		if middleNameTH != nil {
			p.MiddleNameTH = *middleNameTH
		}
		if lastNameTH != nil {
			p.LastNameTH = *lastNameTH
		}
		if firstNameEN != nil {
			p.FirstNameEN = *firstNameEN
		}
		if middleNameEN != nil {
			p.MiddleNameEN = *middleNameEN
		}
		if lastNameEN != nil {
			p.LastNameEN = *lastNameEN
		}
		if dateOfBirth != nil {
			p.DateOfBirth = dateOfBirth
		}
		if gender != nil {
			p.Gender = *gender
		}
		if nationalID != nil {
			p.NationalID = *nationalID
		}
		if passportID != nil {
			p.PassportID = *passportID
		}
		if phoneNumber != nil {
			p.PhoneNumber = *phoneNumber
		}
		if email != nil {
			p.Email = *email
		}

		patients = append(patients, &p)
	}

	if err := rows.Err(); err != nil {
		h.logger.Error("Error reading patient rows", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading patients"})
		return
	}

	h.logger.Info("Patient search completed", "hospital", hospital, "results_count", len(patients))
	c.JSON(http.StatusOK, models.SearchPatientResponse{
		Patient: patients,
	})
}

func (h *Handlers) GetPatientByID(c *gin.Context) {
	hospital, exists := c.Get("hospital")
	if !exists {
		h.logger.Warn("Unauthorized patient retrieval attempt - no hospital in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	identifier := c.Param("id")
	h.logger.Debug("Get patient by identifier request", "identifier", identifier, "hospital", hospital)

	ctx := context.Background()

	// Query patient and verify hospital matches
	query := `
		SELECT id, patient_hn, first_name_th, middle_name_th, last_name_th, first_name_en, middle_name_en, last_name_en, 
		 date_of_birth, gender, national_id, passport_id, phone_number, email 
		FROM patients 
		WHERE national_id = $1 OR passport_id = $1
	`

	var p models.Patient
	var (
		firstNameTH, middleNameTH, lastNameTH              *string
		firstNameEN, middleNameEN, lastNameEN              *string
		gender, nationalID, passportID, phoneNumber, email *string
		dateOfBirth                                        *models.Date
	)

	err := h.db.QueryRow(ctx, query, identifier).Scan(
		&p.ID, &p.PatientHN,
		&firstNameTH, &middleNameTH, &lastNameTH,
		&firstNameEN, &middleNameEN, &lastNameEN,
		&dateOfBirth, &gender,
		&nationalID, &passportID, &phoneNumber, &email,
	)
	if err != nil {
		h.logger.Warn("Patient not found", "identifier", identifier, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	if firstNameTH != nil {
		p.FirstNameTH = *firstNameTH
	}
	if middleNameTH != nil {
		p.MiddleNameTH = *middleNameTH
	}
	if lastNameTH != nil {
		p.LastNameTH = *lastNameTH
	}
	if firstNameEN != nil {
		p.FirstNameEN = *firstNameEN
	}
	if middleNameEN != nil {
		p.MiddleNameEN = *middleNameEN
	}
	if lastNameEN != nil {
		p.LastNameEN = *lastNameEN
	}
	if dateOfBirth != nil {
		p.DateOfBirth = dateOfBirth
	}
	if gender != nil {
		p.Gender = *gender
	}
	if nationalID != nil {
		p.NationalID = *nationalID
	}
	if passportID != nil {
		p.PassportID = *passportID
	}
	if phoneNumber != nil {
		p.PhoneNumber = *phoneNumber
	}
	if email != nil {
		p.Email = *email
	}

	if p.PatientHN != hospital.(string) {
		h.logger.Warn("Access denied - patient belongs to different hospital",
			"identifier", identifier,
			"patient_hospital", p.PatientHN,
			"staff_hospital", hospital,
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied - patient belongs to different hospital"})
		return
	}

	h.logger.Info("Patient retrieved successfully", "identifier", identifier, "hospital", hospital)
	c.JSON(http.StatusOK, p)
}
