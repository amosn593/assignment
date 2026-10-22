package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/amosn593/assignment/internal/database"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	Queries *database.Queries
	DB      *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		Queries: database.New(db),
		DB:      db,
	}
}

type BookAppointmentRequest struct {
	DoctorID int64  `json:"doctor_id" binding:"required"`
	PatientID int64 `json:"patient_id" binding:"required"`
	StartsAt string `json:"starts_at" binding:"required"` // RFC3339 format
}

type CancelRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type RescheduleRequest struct {
	NewStartsAt string `json:"new_starts_at" binding:"required"`
}

// @Summary Book an appointment
// @Description Book a 30-minute slot with a doctor based on configured working hours
// @Accept json
// @Produce json
// @Param request body BookAppointmentRequest true "Booking Details"
// @Success 201 {object} database.Appointment
// @Failure 400 {object} map[string]string
// @Router /appointments [post]
func (h *Handler) BookAppointment(c *gin.Context) {
	var req BookAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or missing fields"})
		return
	}

	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid starts_at format, use RFC3339"})
		return
	}

	// Bonus: Prevention of bookings within 1 hour of now
	if time.Now().Add(1 * time.Hour).After(startsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bookings must be made at least 1 hour in advance"})
		return
	}

	if startsAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot book appointments in the past"})
		return
	}

	// Verify doctor exists
	_, err = h.Queries.GetDoctor(c.Request.Context(), req.DoctorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Doctor not found"})
		return
	}

	// Verify patient exists (referential integrity enforced)
	_, err = h.Queries.GetPatient(c.Request.Context(), req.PatientID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Patient not found"})
		return
	}

	// Fixed duration of 30 minutes
	endsAt := startsAt.Add(30 * time.Minute)

	// Validate against DoctorWorkingHours for that specific weekday
	weekday := int64(startsAt.Weekday())
	workingHours, err := h.Queries.GetDoctorWorkingHoursForDay(c.Request.Context(), database.GetDoctorWorkingHoursForDayParams{
		DoctorID: req.DoctorID,
		Weekday:  weekday,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Doctor does not work on this day"})
		return
	}

	if !h.isWithinWorkingHours(workingHours, startsAt, endsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment falls outside doctor working hours"})
		return
	}

	// Check if slot is already taken
	dayStart := time.Date(startsAt.Year(), startsAt.Month(), startsAt.Day(), 0, 0, 0, 0, startsAt.Location())
	dayEnd := dayStart.Add(24 * time.Hour)

	existing, err := h.Queries.GetAppointmentsForDoctorDate(c.Request.Context(), database.GetAppointmentsForDoctorDateParams{
		DoctorID:   req.DoctorID,
		Starts_At:  dayStart.Format("2006-01-02 15:04:05"),
		Starts_At2: dayEnd.Format("2006-01-02 15:04:05"),
	})
	if err == nil {
		for _, appt := range existing {
			apptStart, _ := time.Parse("2006-01-02 15:04:05", appt.StartsAt)
			if apptStart.Equal(startsAt) {
				c.JSON(http.StatusConflict, gin.H{"error": "Time slot is already taken"})
				return
			}
		}
	}

	appt, err := h.Queries.CreateAppointment(c.Request.Context(), database.CreateAppointmentParams{
		PatientID: req.PatientID,
		DoctorID:  req.DoctorID,
		StartsAt:  startsAt.Format("2006-01-02 15:04:05"),
		EndsAt:    endsAt.Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create appointment"})
		return
	}

	c.JSON(http.StatusCreated, appt)
}

// @Summary Get doctor availability
// @Description Return all available 30-minute slots for a doctor on a given date (YYYY-MM-DD)
// @Produce json
// @Param id path int true "Doctor ID"
// @Param date query string true "Date in YYYY-MM-DD format"
// @Success 200 {array} string
// @Router /doctors/{id}/availability [get]
func (h *Handler) GetDoctorAvailability(c *gin.Context) {
	docID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid doctor ID"})
		return
	}

	dateStr := c.Query("date")
	targetDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format, use YYYY-MM-DD"})
		return
	}

	weekday := int64(targetDate.Weekday())
	workingHours, err := h.Queries.GetDoctorWorkingHoursForDay(c.Request.Context(), database.GetDoctorWorkingHoursForDayParams{
		DoctorID: docID,
		Weekday:  weekday,
	})
	if err != nil {
		// Doctor does not work on this day
		c.JSON(http.StatusOK, []string{})
		return
	}

	startTimeParts, _ := time.Parse("15:04", workingHours.StartTime)
	endTimeParts, _ := time.Parse("15:04", workingHours.EndTime)

	slotStart := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), startTimeParts.Hour(), startTimeParts.Minute(), 0, 0, targetDate.Location())
	dayEnd := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), endTimeParts.Hour(), endTimeParts.Minute(), 0, 0, targetDate.Location())

	dayStartStr := slotStart.Format("2006-01-02 00:00:00")
	dayEndStr := slotStart.Add(24 * time.Hour).Format("2006-01-02 00:00:00")

	bookedApps, _ := h.Queries.GetAppointmentsForDoctorDate(c.Request.Context(), database.GetAppointmentsForDoctorDateParams{
		DoctorID:   docID,
		Starts_At:  dayStartStr,
		Starts_At2: dayEndStr,
	})

	bookedMap := make(map[string]bool)
	for _, appt := range bookedApps {
		bookedMap[appt.StartsAt] = true
	}

	var availableSlots []string
	// Generate 30-minute intervals ending no later than closing time
	for slotStart.Add(30 * time.Minute).Before(dayEnd) || slotStart.Add(30 * time.Minute).Equal(dayEnd) {
		slotStr := slotStart.Format("2006-01-02 15:04:05")
		if !bookedMap[slotStr] && slotStart.After(time.Now()) {
			availableSlots = append(availableSlots, slotStart.Format(time.RFC3339))
		}
		slotStart = slotStart.Add(30 * time.Minute)
	}

	c.JSON(http.StatusOK, availableSlots)
}

// @Summary Cancel an appointment
// @Description Cancel an existing appointment with a reason, keeping the record for history
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param request body CancelRequest true "Cancellation Reason"
// @Success 200 {object} database.Appointment
// @Router /appointments/{id}/cancel [patch]
func (h *Handler) CancelAppointment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cancellation reason is required"})
		return
	}

	appt, err := h.Queries.GetAppointmentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	if appt.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Appointment is already cancelled"})
		return
	}

	updated, err := h.Queries.CancelAppointment(c.Request.Context(), database.CancelAppointmentParams{
		CancellationReason: sql.NullString{String: req.Reason, Valid: true},
		ID:                 id,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel appointment"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// @Summary Reschedule an appointment
// @Description Move an appointment to a new slot
// @Accept json
// @Produce json
// @Param id path int true "Appointment ID"
// @Param request body RescheduleRequest true "New Slot Details"
// @Success 200 {object} database.Appointment
// @Router /appointments/{id}/reschedule [patch]
func (h *Handler) RescheduleAppointment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment ID"})
		return
	}

	var req RescheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	appt, err := h.Queries.GetAppointmentByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Appointment not found"})
		return
	}

	if appt.Status == "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot reschedule a cancelled appointment"})
		return
	}

	newStartsAt, err := time.Parse(time.RFC3339, req.NewStartsAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid new_starts_at format"})
		return
	}

	newEndsAt := newStartsAt.Add(30 * time.Minute)
	weekday := int64(newStartsAt.Weekday())
	workingHours, err := h.Queries.GetDoctorWorkingHoursForDay(c.Request.Context(), database.GetDoctorWorkingHoursForDayParams{
		DoctorID: appt.DoctorID,
		Weekday:  weekday,
	})
	if err != nil || !h.isWithinWorkingHours(workingHours, newStartsAt, newEndsAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New slot falls outside doctor working hours"})
		return
	}

	// Mark old appointment as rescheduled/cancelled
	_, _ = h.Queries.UpdateAppointmentStatus(c.Request.Context(), database.UpdateAppointmentStatusParams{
		Status: "rescheduled",
		ID:     appt.ID,
	})

	// Create new booking
	newAppt, err := h.Queries.CreateAppointment(c.Request.Context(), database.CreateAppointmentParams{
		PatientID: appt.PatientID,
		DoctorID:  appt.DoctorID,
		StartsAt:  newStartsAt.Format("2006-01-02 15:04:05"),
		EndsAt:    newEndsAt.Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reschedule appointment"})
		return
	}

	c.JSON(http.StatusOK, newAppt)
}

// @Summary Get patient upcoming appointments
// @Description Return upcoming appointments for a patient sorted by date
// @Produce json
// @Param id path int true "Patient ID"
// @Success 200 {array} database.Appointment
// @Router /patients/{id}/appointments [get]
func (h *Handler) GetPatientAppointments(c *gin.Context) {
	patientID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid patient ID"})
		return
	}

	appts, err := h.Queries.GetUpcomingAppointmentsForPatient(c.Request.Context(), patientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch appointments"})
		return
	}

	c.JSON(http.StatusOK, appts)
}

func (h *Handler) isWithinWorkingHours(wh database.DoctorWorkingHour, start, end time.Time) bool {
	sParts, _ := time.Parse("15:04", wh.StartTime)
	eParts, _ := time.Parse("15:04", wh.EndTime)

	dayStart := time.Date(start.Year(), start.Month(), start.Day(), sParts.Hour(), sParts.Minute(), 0, 0, start.Location())
	dayEnd := time.Date(start.Year(), start.Month(), start.Day(), eParts.Hour(), eParts.Minute(), 0, 0, start.Location())

	return (start.Equal(dayStart) || start.After(dayStart)) && (end.Equal(dayEnd) || end.Before(dayEnd))
}