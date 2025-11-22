package handlers

import (
	"agnos_demo/internal/middleware"
	"agnos_demo/internal/mocks"
	"agnos_demo/internal/models"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func setupRouter(h *Handlers) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/health", h.HealthCheck)
	r.POST("/staff/create", h.CreateStaff)
	r.POST("/staff/login", h.LoginStaff)

	protected := r.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/patient/search", h.SearchPatient)
		protected.GET("/patient/search/:id", h.GetPatientByID)
	}
	return r
}

func TestHealthCheck(t *testing.T) {
	mockDB := new(mocks.MockDB)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	h := NewHandlers(mockDB, logger)
	r := setupRouter(h)

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "OK")
}

func TestCreateStaff(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Success", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		// Setup expectations
		testID := uuid.New()
		mockRow.On("Scan", mock.Anything).Run(func(args mock.Arguments) {
			// Set the UUID in the first argument
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "testuser", "password": "password", "hospital": "hn-001"}`
		req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Contains(t, w.Body.String(), "Staff created successfully")
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Missing Fields", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "testuser"}`
		req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Database Error", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		mockRow.On("Scan", mock.Anything).Return(errors.New("database error"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "testuser", "password": "password", "hospital": "hn-001"}`
		req, _ := http.NewRequest("POST", "/staff/create", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestLoginStaff(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("Success", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		testID := uuid.New()
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
			if username, ok := args.Get(1).(*string); ok {
				*username = "loginuser"
			}
			if passwordHash, ok := args.Get(2).(*string); ok {
				*passwordHash = string(hashedPassword)
			}
			if hospital, ok := args.Get(3).(*string); ok {
				*hospital = "hn-001"
			}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "loginuser", "password": "password123", "hospital": "hn-001"}`
		req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "token")
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("no rows"))
		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "nonexistent", "password": "password123", "hospital": "hn-001"}`
		req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		testID := uuid.New()
		mockRow.On("Scan", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
			if username, ok := args.Get(1).(*string); ok {
				*username = "loginuser"
			}
			if passwordHash, ok := args.Get(2).(*string); ok {
				*passwordHash = string(hashedPassword)
			}
			if hospital, ok := args.Get(3).(*string); ok {
				*hospital = "hn-001"
			}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		body := `{"username": "loginuser", "password": "wrongpassword", "hospital": "hn-001"}`
		req, _ := http.NewRequest("POST", "/staff/login", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}

func TestSearchPatient(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Search Success Empty", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRows := new(mocks.MockRows)

		// Setup expectations - no rows returned
		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return()

		mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search?first_name=John", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response models.SearchPatientResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response.Patient, 0)

		mockDB.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})

	t.Run("Search Success With Data", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRows := new(mocks.MockRows)

		// Setup expectations - 1 row returned
		mockRows.On("Next").Return(true).Once()
		mockRows.On("Next").Return(false).Once()

		testID := uuid.New()
		mockRows.On("Scan",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Run(func(args mock.Arguments) {
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
			if patientHN, ok := args.Get(1).(*string); ok {
				*patientHN = "hn-001"
			}
			// Simulate nullable fields being populated
			if firstNameEN, ok := args.Get(5).(**string); ok {
				s := "John"
				*firstNameEN = &s
			}
			if dob, ok := args.Get(8).(**models.Date); ok {
				d := models.Date{Time: time.Date(1985, 3, 20, 0, 0, 0, 0, time.UTC)}
				*dob = &d
			}
		}).Return(nil)

		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return()

		mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search?first_name=John", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify JSON string directly to check date format
		assert.Contains(t, w.Body.String(), `"date_of_birth":"1985-03-20"`)

		var response models.SearchPatientResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Len(t, response.Patient, 1)
		assert.Equal(t, testID, response.Patient[0].ID)
		assert.Equal(t, "John", response.Patient[0].FirstNameEN)

		mockDB.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})

	t.Run("Database Query Error", func(t *testing.T) {
		mockDB := new(mocks.MockDB)

		mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Rows Error", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRows := new(mocks.MockRows)

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(errors.New("rows error"))
		mockRows.On("Close").Return()

		mockDB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("Search With All Filters", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRows := new(mocks.MockRows)

		mockRows.On("Next").Return(false)
		mockRows.On("Err").Return(nil)
		mockRows.On("Close").Return()

		// Expect query with many arguments
		mockDB.On("Query", mock.Anything, mock.Anything,
			mock.Anything, // hospital prefix
			mock.Anything, // patient_hn
			mock.Anything, // national_id
			mock.Anything, // passport_id
			mock.Anything, // first_name
			mock.Anything, // middle_name
			mock.Anything, // last_name
			mock.Anything, // dob
		).Return(mockRows, nil)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		// Construct URL with all parameters
		url := "/patient/search?patient_hn=hn-001-01&national_id=123&passport_id=A123&first_name=John&middle_name=M&last_name=Doe&date_of_birth=1980-01-01"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
		mockRows.AssertExpectations(t)
	})

	t.Run("Unauthorized", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		req, _ := http.NewRequest("GET", "/patient/search", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetPatientByID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	t.Run("Success", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		testID := uuid.New()
		mockRow.On("Scan",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Run(func(args mock.Arguments) {
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
			if patientHN, ok := args.Get(1).(*string); ok {
				*patientHN = "hn-001"
			}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search/1234567890123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		mockRow.On("Scan",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Return(errors.New("no rows"))

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search/1234567890123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})

	t.Run("Different Hospital", func(t *testing.T) {
		mockDB := new(mocks.MockDB)
		mockRow := new(mocks.MockRow)

		testID := uuid.New()
		mockRow.On("Scan",
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		).Run(func(args mock.Arguments) {
			if id, ok := args.Get(0).(*uuid.UUID); ok {
				*id = testID
			}
			if patientHN, ok := args.Get(1).(*string); ok {
				*patientHN = "hn-002" // Different hospital
			}
		}).Return(nil)

		mockDB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)

		h := NewHandlers(mockDB, logger)
		r := setupRouter(h)

		token, _ := middleware.GenerateToken(uuid.New().String(), "hn-001")

		req, _ := http.NewRequest("GET", "/patient/search/1234567890123", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockDB.AssertExpectations(t)
		mockRow.AssertExpectations(t)
	})
}
