
package handler_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amosn593/assignment/internal/router"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test db: %v", err)
	}

	schema := `
	CREATE TABLE doctors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		specialty TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE doctor_working_hours (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		doctor_id INTEGER NOT NULL,
		weekday INTEGER NOT NULL,
		start_time TEXT NOT NULL,
		end_time TEXT NOT NULL
	);
	CREATE TABLE patients (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		full_name TEXT NOT NULL,
		email TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE appointments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		patient_id INTEGER NOT NULL,
		doctor_id INTEGER NOT NULL,
		starts_at DATETIME NOT NULL,
		ends_at DATETIME NOT NULL,
		status TEXT NOT NULL DEFAULT 'booked',
		cancellation_reason TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	INSERT INTO doctors (id, full_name, specialty) VALUES (1, 'Dr. Test', 'General');
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	// Insert working hours for ALL days of the week (0 to 6) to prevent weekday mismatches
	for w := 0; w <= 6; w++ {
		_, err = db.Exec(`INSERT INTO doctor_working_hours (doctor_id, weekday, start_time, end_time) VALUES (1, ?, '00:00', '23:59')`, w)
		if err != nil {
			t.Fatalf("Failed to insert working hours: %v", err)
		}
	}

	_, err = db.Exec(`INSERT INTO patients (id, full_name, email) VALUES (1, 'Test Patient', 'test@example.com')`)
	if err != nil {
		t.Fatalf("Failed to insert patient: %v", err)
	}

	return db
}

func TestBookAndCancelAppointment(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	r := router.SetupRouter(db)

	futureTime := time.Now().Add(48 * time.Hour).Truncate(time.Hour).Add(10 * time.Hour)
	reqBody := map[string]interface{}{
		"doctor_id":  1,
		"patient_id": 1,
		"starts_at":  futureTime.Format(time.RFC3339),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	// 1. Test Booking
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/appointments", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Test Availability Check
	wAvail := httptest.NewRecorder()
	reqAvail, _ := http.NewRequest("GET", "/doctors/1/availability?date="+futureTime.Format("2006-01-02"), nil)
	r.ServeHTTP(wAvail, reqAvail)
	if wAvail.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on availability, got %d", wAvail.Code)
	}

	// 3. Test Cancellation
	cancelBody, _ := json.Marshal(map[string]string{"reason": "Personal emergency"})
	wCancel := httptest.NewRecorder()
	reqCancel, _ := http.NewRequest("PATCH", "/appointments/1/cancel", bytes.NewBuffer(cancelBody))
	reqCancel.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wCancel, reqCancel)

	if wCancel.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on cancel, got %d: %s", wCancel.Code, wCancel.Body.String())
	}
}