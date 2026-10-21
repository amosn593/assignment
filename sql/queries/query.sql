-- name: GetDoctor :one
SELECT * FROM doctors WHERE id = ?;

-- name: GetDoctorWorkingHoursForDay :one
SELECT * FROM doctor_working_hours WHERE doctor_id = ? AND weekday = ?;

-- name: GetPatient :one
SELECT * FROM patients WHERE id = ?;

-- name: GetAppointmentsForDoctorDate :many
SELECT * FROM appointments 
WHERE doctor_id = ? 
  AND status = 'booked'
  AND datetime(starts_at) >= datetime(?) 
  AND datetime(starts_at) < datetime(?);

-- name: CreateAppointment :one
INSERT INTO appointments (patient_id, doctor_id, starts_at, ends_at, status)
VALUES (?, ?, ?, ?, 'booked')
RETURNING *;

-- name: GetAppointmentByID :one
SELECT * FROM appointments WHERE id = ?;

-- name: CancelAppointment :one
UPDATE appointments 
SET status = 'cancelled', cancellation_reason = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? 
RETURNING *;

-- name: UpdateAppointmentStatus :one
UPDATE appointments 
SET status = ?, updated_at = CURRENT_TIMESTAMP 
WHERE id = ? 
RETURNING *;

-- name: GetUpcomingAppointmentsForPatient :many
SELECT * FROM appointments 
WHERE patient_id = ? 
  AND status = 'booked' 
  AND datetime(starts_at) >= datetime('now')
ORDER BY starts_at ASC;