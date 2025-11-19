package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Machine struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;" json:"id"`
	SellerID            uint           `gorm:"not null" json:"seller_id"` // Changed to uint to match Auth service User ID
	Title               string         `gorm:"type:varchar(100);not null" json:"title"`
	Description         string         `gorm:"type:text;not null" json:"description"`
	Manufacturer        string         `gorm:"type:varchar(100);not null" json:"manufacturer"`
	ModelNumber         string         `gorm:"type:varchar(100);not null" json:"model_number"`
	YearOfManufacture   int            `gorm:"type:int;not null" json:"year_of_manufacture"`
	Status              string         `gorm:"type:varchar(50);default:'pending_inspection'" json:"status"` // pending_inspection, listed, sold, rented
	ListingType         string         `gorm:"type:varchar(50);not null" json:"listing_type"`               // sale, rent, both
	PriceForSale        float64        `gorm:"type:decimal(10,2);default:0" json:"price_for_sale"`
	RentalPricePerDay   float64        `gorm:"type:decimal(10,2);default:0" json:"rental_price_per_day"`
	RentalPricePerMonth float64        `gorm:"type:decimal(10,2);default:0" json:"rental_price_per_month"`
	SecurityDeposit     float64        `gorm:"type:decimal(10,2);default:0" json:"security_deposit"`
	Specs               datatypes.JSON `gorm:"type:jsonb" json:"specs"` // Flexible JSON for technical specs
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID
func (m *Machine) BeforeCreate(tx *gorm.DB) (err error) {
	m.ID = uuid.New()
	return
}