package dto

type PathID struct {
	ID     string `uri:"id" binding:"required,uuid"`
	CampID string `uri:"campId" binding:"omitempty,uuid"`
}
type QueryClient struct {
	ClientID string `form:"client_id" json:"client_id" binding:"required,uuid"`
}
type TimeSetBody struct {
	CurrentDate *uint `json:"current_date" binding:"required"`
}
type AdUser struct {
	AdID              string  `json:"ad_id" gorm:"column:campaign_id"`
	AdTitle           string  `json:"ad_title"`
	AdText            string  `json:"ad_text"`
	AdvertiserID      string  `json:"advertiser_id"`
	Score             float64 `json:"-"`
	CostPerImpression float64 `json:"-"`
}
type Stats struct {
	ImpressionsCount int     `json:"impressions_count"`
	ClicksCount      int     `json:"clicks_count"`
	Conversion       float64 `json:"conversion"`
	SpentImpressions float64 `json:"spent_impressions"`
	SpentClicks      float64 `json:"spent_clicks"`
	SpentTotal       float64 `json:"spent_total"`
}
type DailyStats struct {
	Stats
	Date int64 `json:"date"`
}
type ModerationBody struct {
	Moderation *bool `json:"moderation" binding:"required"`
}
type ModerationVerdictBody struct {
	Verdict *bool `json:"verdict" binding:"required"`
}
type NeuroBody struct {
	AdTitle string `json:"ad_title" binding:"required"`
}
