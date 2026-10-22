# Project assignment

# Section One - System Design

I have used Go BluePrint to generate initial project boilerplate code.

I have also use cursor as my coding assistant.

The api is deployed in render cloud.

#### Models

Doctor

Represents a doctor who can receive appointments.

Fields:

id: Unique doctor identifier.

full_name: Doctor's full name.

specialty: Medical specialty.

created_at: Record creation timestamp.

Relationships:

A doctor can have multiple working-hour records.

A doctor can have multiple appointments.

DoctorWorkingHours

Defines when a doctor is available for appointments.

Fields:

id: Unique working-hours record identifier.

doctor_id: Reference to the doctor.

weekday: Day of the week.

start_time: Start of the working period.

end_time: End of the working period.

Patient

Represents a patient who books appointments.

Fields:

id: Unique patient identifier.

full_name: Patient's full name.

email: Patient's email address, unique.

created_at: Record creation timestamp.

Appointment

Represents a booking between a patient and a doctor.

Fields:

id: Unique appointment identifier.

patient_id: Reference to the patient.

doctor_id: Reference to the doctor.

starts_at: Appointment start timestamp.

ends_at: Appointment end timestamp.

status: Appointment status, either booked or cancelled.

cancellation_reason: Reason for cancellation, when applicable.

created_at: Record creation timestamp.

updated_at: Last modification timestamp.

Relationships:

Each appointment belongs to one patient.

Each appointment belongs to one doctor.

#### Components

##### HTTP/API Layer/Router

Responsible for receiving HTTP requests, parsing JSON, validating request formats, invoking application services, and returning JSON responses.
`
Seed Initial Data
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
(1, 'John Doe', 'john@example.com')
`

##### Booking Service/Handler

Contains the core business rules for appointment management.

Responsibilities:

Validate doctor and patient identifiers.

Validate appointment timestamps.

Ensure appointments follow 30-minute slot boundaries.

Check doctor working hours.

Enforce the one-hour advance-booking rule.

Generate available appointment slots.

Coordinate booking, cancellation, and rescheduling operations.

Translate business validation failures into meaningful errors.

##### Data Access/Database

SQL queries are maintained separately from business logic. sqlc is intended to generate type-safe database access code from these queries, while golang-migrate manages schema changes.

Database constraints provide a final safeguard against invalid relationships and duplicate active bookings.

#### Key Decisions and trade-offs

1. SQLite Database instead of a client-server database
   SQLite keeps the initial application easy to install and run without managing a separate database server.

The trade-off is that SQLite is less suitable for high write concurrency and multi-instance deployments than a server database. If the clinic grows significantly, PostgreSQL would be a reasonable migration target.

2. No authentication  
   Authentication is intentionally omitted as requested. The API accepts identifiers supplied by the caller.

This simplifies the assessment implementation but means the service must not be treated as a secure production patient-data system. A production version should introduce authentication, authorization, and patient-level access controls.

# Section Two - API IMPLEMENTATION

Created all the required apis using Go gin framework

# Section Three - DEPLOYMENT & CI/CD

- Public URL
- Main branch, triggers deployment on push or merging a pull request.
- Deployments are on main branch on push and pull request.
- CI process: Installs dependencies, builds the app, then run tests.
- Deployment process: on push and pull request and after successful ci actions runs, pipeline builds a production app binary, Logins in to azure using store app secrets, then deploys the app to azure app service.

# Section Four - AI Reflection

- I used AI to generate most of the code, by providing detailed and task specific prompts.
- Prompt: "Write a Go function using time.Parse that generates 30-minute intervals between a doctor's start and end times while excluding booked slots and past hours.", The AI correctly handled timezone parsing constraints and gave a robust iteration loop over dayStart to dayEnd with 30-minute increments.
- The initial AI output for date comparison in SQL used improper SQLite date string formats (DATETIME comparison without correct formatting). I corrected by checking with SQLite string formatting rules and standardizing timestamps to "2006-01-02 15:04:05"
- Database choice(sqlite) and no authentication decision was made without use of AI
