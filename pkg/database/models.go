// pkg/database/models.go
package database

import (
	"time"

	"gorm.io/gorm"
)

// User representa al usuario del sistema
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null" json:"username"`
	Password  string         `gorm:"not null" json:"-"` // Ocultar en JSON
	Email     string         `gorm:"uniqueIndex" json:"email"`
	FullName  string         `json:"full_name"`
	Role      string         `gorm:"default:'operator'" json:"role"`
	Jobs      []Job          `gorm:"foreignKey:CreatedByID" json:"jobs,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// JobStatus define los estados del flujo de trabajo
type JobStatus string

const (
	StatusPending   JobStatus = "PENDING"
	StatusDone      JobStatus = "DONE"
	StatusExecuted  JobStatus = "EXECUTED"
)

// Job representa un trabajo realizado por un usuario
type Job struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	Status      JobStatus      `gorm:"default:'PENDING'" json:"status"`
	B3          string         `gorm:"index" json:"b3"` // Estación asociada
	CreatedByID uint           `json:"created_by_id"`
	CreatedBy   User           `gorm:"foreignKey:CreatedByID" json:"created_by"`
	Comments    []Comment      `json:"comments,omitempty"`
	History     []JobHistory   `json:"history,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Comment representa un comentario en un Job
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	JobID     uint      `gorm:"index" json:"job_id"`
	UserID    uint      `json:"user_id"`
	User      User      `json:"user"`
	Content   string    `gorm:"not null" json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// JobHistory registra cada cambio de estado o intervención
type JobHistory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	JobID       uint      `gorm:"index" json:"job_id"`
	UserID      uint      `json:"user_id"`
	OldStatus   JobStatus `json:"old_status"`
	NewStatus   JobStatus `json:"new_status"`
	Action      string    `json:"action"` // e.g., "Updated status", "Generated XML"
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
