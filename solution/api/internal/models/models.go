package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type Client struct {
	ClientID uuid.UUID `gorm:"primaryKey;type:uuid" json:"client_id" binding:"required,uuid"`
	Login    string    `gorm:"unique" json:"login" binding:"required"`
	Age      uint      `json:"age" binding:"required"`
	Location string    `json:"location" binding:"required"`
	Gender   string    `json:"gender" binding:"required,oneof=MALE FEMALE"`
}

type Advertiser struct {
	AdvertiserID uuid.UUID `gorm:"primaryKey;type:uuid" json:"advertiser_id" binding:"required,uuid"`
	Name         string    `json:"name" binding:"required"`
}

type Campaign struct {
	CampaignID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"campaign_id" binding:"-"`
	AdvertiserID      uuid.UUID  `gorm:"type:uuid" json:"advertiser_id" binding:"-"`
	Advertiser        Advertiser `json:"-" binding:"-"`
	ImpressionsLimit  *uint      `json:"impressions_limit" binding:"required,gtefield=ClicksLimit"`
	ClicksLimit       *uint      `json:"clicks_limit" binding:"required"`
	CostPerImpression *float64   `json:"cost_per_impression" binding:"required"`
	CostPerClick      *float64   `json:"cost_per_click" binding:"required"`
	AdTitle           string     `json:"ad_title" binding:"required"`
	AdText            string     `json:"ad_text" binding:"required"`
	StartDate         *uint      `json:"start_date" binding:"required"`
	EndDate           *uint      `json:"end_date" binding:"required,gtefield=StartDate"`
	Targeting         Targeting  `gorm:"type:jsonb" json:"targeting"`
	Image             *string    `json:"image" binding:"-"`
	Moderated         bool       `json:"moderated" binding:"-"`
}
type CampaignUpdate struct {
	CampaignID        uuid.UUID  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"campaign_id" binding:"-"`
	AdvertiserID      uuid.UUID  `gorm:"type:uuid" json:"advertiser_id" binding:"-"`
	Advertiser        Advertiser `json:"-" binding:"-"`
	ImpressionsLimit  *uint      `json:"impressions_limit" binding:"required,gtefield=ClicksLimit"`
	ClicksLimit       *uint      `json:"clicks_limit" binding:"required"`
	CostPerImpression *float64   `json:"cost_per_impression"`
	CostPerClick      *float64   `json:"cost_per_click"`
	AdTitle           string     `json:"ad_title" binding:"required"`
	AdText            string     `json:"ad_text" binding:"required"`
	Targeting         Targeting  `gorm:"type:jsonb" json:"targeting"`
}
type Targeting struct {
	Gender   *string `json:"gender" binding:"omitempty,oneof=MALE FEMALE ALL"`
	AgeFrom  *uint   `json:"age_from" binding:"omitempty"`
	AgeTo    *uint   `json:"age_to" binding:"omitempty,gtenrfield=AgeFrom"`
	Location *string `json:"location" binding:"omitempty"`
}

func (t *Targeting) Scan(value interface{}) error {
	if value == nil {
		*t = Targeting{}
		return nil
	}
	data, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Targeting")
	}
	return json.Unmarshal(data, t)
}

func (t Targeting) Value() (driver.Value, error) {
	return json.Marshal(t)
}

type Event struct {
	ID         uint
	ClientID   uuid.UUID `gorm:"type:uuid"`
	Client     Client
	CampaignID uuid.UUID `gorm:"type:uuid"`
	Campaign   Campaign
	Type       string
	Cost       float64
	Day        uint
}

type MLScore struct {
	// составной первичный ключ
	ClientID     uuid.UUID `gorm:"primaryKey;type:uuid" json:"client_id" binding:"required,uuid"`
	AdvertiserID uuid.UUID `gorm:"primaryKey;type:uuid" json:"advertiser_id" binding:"required,uuid"`
	Score        *uint     `json:"score" binding:"required"`

	Client     Client     `json:"-" binding:"-"`
	Advertiser Advertiser `json:"-" binding:"-"`
}

type Setting struct {
	ID         uint
	Day        uint
	Moderation bool
}
