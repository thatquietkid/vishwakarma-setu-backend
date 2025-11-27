package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type InspectionReport struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	MachineID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"machine_id"`
	InspectorID uint           `gorm:"not null" json:"inspector_id"`
	ReportType  string         `gorm:"type:varchar(50);default:'listing'" json:"report_type"`

	InspectionDate time.Time      `json:"inspection_date"`
	Verdict        string         `gorm:"type:varchar(50)" json:"verdict"`
	Summary        string         `gorm:"type:text" json:"summary"`

	// FIX: Added swaggertype:"object"
	ReportData     datatypes.JSON `gorm:"type:jsonb" json:"report_data" swaggertype:"object"`

	// FIX: Added swaggertype:"array,string"
	MediaURLs      datatypes.JSON `gorm:"type:jsonb" json:"media_urls" swaggertype:"array,string"`

	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *InspectionReport) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}