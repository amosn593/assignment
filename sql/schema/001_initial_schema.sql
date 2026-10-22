CREATE TABLE doctors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    specialty TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE doctor_working_hours (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    doctor_id INTEGER NOT NULL,
    weekday INTEGER NOT NULL, -- 0 (Sunday) to 6 (Saturday)
    start_time TEXT NOT NULL, -- Format "09:00"
    end_time TEXT NOT NULL,   -- Format "17:00"
    FOREIGN KEY (doctor_id) REFERENCES doctors(id) ON DELETE CASCADE
);

CREATE TABLE patients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    full_name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE appointments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    patient_id INTEGER NOT NULL,
    doctor_id INTEGER NOT NULL,
    starts_at DATETIME NOT NULL,
    ends_at DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'booked', -- booked, cancelled, rescheduled
    cancellation_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (patient_id) REFERENCES patients(id),
    FOREIGN KEY (doctor_id) REFERENCES doctors(id)
);

-- Seed Initial Data
INSERT INTO doctors (id, full_name, specialty) VALUES 
(1, 'Dr. Alice Smith', 'General Practitioner'),
(2, 'Dr. Bob Jones', 'Pediatrics');

-- Working hours for Dr. Alice Smith (Mon-Fri 09:00 to 17:00)
INSERT INTO doctor_working_hours (doctor_id, weekday, start_time, end_time) VALUES 
(1, 1, '09:00', '17:00'),
(1, 2, '09:00', '17:00'),
(1, 3, '09:00', '17:00'),
(1, 4, '09:00', '17:00'),
(1, 5, '09:00', '17:00');

INSERT INTO patients (id, full_name, email) VALUES 
(1, 'John Doe', 'john@example.com');