package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Rental struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	MachineID           uuid.UUID      `gorm:"type:uuid;not null" json:"machine_id"`
	RenterID            uint           `gorm:"not null" json:"renter_id"` // Matches Auth Service ID type
	
	StartDate           time.Time      `gorm:"not null" json:"start_date"`
	EndDate             time.Time      `gorm:"not null" json:"end_date"`
	
	// Financials
	TotalAmount         float64        `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	SecurityDeposit     float64        `gorm:"type:decimal(10,2);not null" json:"security_deposit"`
	PlatformFee         float64        `gorm:"type:decimal(10,2);default:0" json:"platform_fee"` // e.g. 5%
	
	// Status Flow: pending -> approved -> active -> completed (or rejected/cancelled)
	Status              string         `gorm:"type:varchar(50);default:'pending'" json:"status"`
	
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`

	// Preloads (Optional, for returning full details)
	Machine             Machine        `gorm:"foreignKey:MachineID" json:"machine,omitempty"`
}

func (r *Rental) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return
}