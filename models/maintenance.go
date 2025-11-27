package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MaintenanceRecord struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	MachineID uuid.UUID `gorm:"type:uuid;not null;index" json:"machine_id"`

	ServiceDate time.Time `json:"service_date"`
	Type        string    `gorm:"type:varchar(50)" json:"type"` // e.g., "Routine", "Repair", "Replacement"
	Description string    `gorm:"type:text" json:"description"`
	Cost        float64   `gorm:"type:decimal(10,2)" json:"cost"`
	Technician  string    `gorm:"type:varchar(100)" json:"technician"`

	// Optional: Link to invoice/document image
	DocumentURL string `gorm:"type:varchar(255)" json:"document_url"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *MaintenanceRecord) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}
