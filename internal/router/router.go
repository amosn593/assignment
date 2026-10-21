package router

import (
	"database/sql"

	"github.com/amosn593/assignment/internal/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()
	h := handler.NewHandler(db)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/appointments", h.BookAppointment)
	r.GET("/doctors/:id/availability", h.GetDoctorAvailability)
	r.PATCH("/appointments/:id/cancel", h.CancelAppointment)
	r.PATCH("/appointments/:id/reschedule", h.RescheduleAppointment)
	r.GET("/patients/:id/appointments", h.GetPatientAppointments)

	return r
}