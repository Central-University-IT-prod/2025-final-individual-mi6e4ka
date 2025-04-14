package services

import (
	"log"

	"git.mi6e4ka.dev/prod-2025/internal/dto"
	"git.mi6e4ka.dev/prod-2025/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AdService struct {
	db *gorm.DB
}

func NewAdService(db *gorm.DB) *AdService {
	return &AdService{db: db}
}

func (s *AdService) GetAds(clientID uuid.UUID, day uint, moderation bool) (*dto.AdUser, error) {
	var client models.Client
	if err := s.db.First(&client, "client_id = ?", clientID).Error; err != nil {
		return nil, err
	}

	// ML нейросетевой подбор алгоритм нейро гипер мега

	// уффф короче это все переписала нейронка, но по моему алгоритму
	var campaignsWithScores dto.AdUser
	subQueryStats := s.db.Table("campaigns").Select(`
		AVG(cost_per_impression) AS avg_cpi,
		STDDEV(cost_per_impression) AS stddev_cpi,
		AVG(cost_per_click) AS avg_cpc,
		STDDEV(cost_per_click) AS stddev_cpc
	`)

	subQueryMLStats := s.db.Table("ml_scores").Select(`
		AVG(score) AS avg_ml,
		STDDEV(score) AS stddev_ml
	`)

	subQueryEvents := s.db.Table("events").Select(`
		campaign_id,
		SUM(CASE WHEN type = 'view' AND client_id = ? THEN 1 ELSE 0 END) AS user_impressions_count,
		SUM(CASE WHEN type = 'view' THEN 1 ELSE 0 END) AS impressions_count,
		SUM(CASE WHEN type = 'click' THEN 1 ELSE 0 END) AS clicks_count
	`, clientID).Where("type IN ('view', 'click')").Group("campaign_id")

	req := s.db.Table("campaigns c").
		Select(`c.*,
			(0.5 * ((c.cost_per_impression - gs.avg_cpi) / COALESCE(NULLIF(gs.stddev_cpi, 0), 1)) +
			0.3 * ((c.cost_per_click - gs.avg_cpc) / COALESCE(NULLIF(gs.stddev_cpc, 0),1)) +
			0.2 * ((COALESCE(m.score,0) - COALESCE(gms.avg_ml,0)) / COALESCE(NULLIF(gms.stddev_ml, 0),1))) *
			(1 - (COALESCE(e.impressions_count, 0) / c.impressions_limit)^3) AS score
		`).
		Joins("LEFT JOIN (?) AS gs ON 1=1", subQueryStats).
		Joins("LEFT JOIN (?) AS gms ON 1=1", subQueryMLStats).
		Joins("LEFT JOIN ml_scores m ON m.advertiser_id = c.advertiser_id AND m.client_id = ?", clientID).
		Joins("LEFT JOIN (?) AS e ON e.campaign_id = c.campaign_id", subQueryEvents).
		Where(`
			COALESCE(e.impressions_count, 0) + 1 < c.impressions_limit * 1.05
			AND COALESCE(e.user_impressions_count, 0) = 0
			AND (c.start_date <= ? AND c.end_date >= ?)`+ /*<-- нет и вот прикиньте я на фикс этого час потратил...*/ `
			AND ((c.targeting->>'gender' IS NULL OR c.targeting->>'gender' = 'ALL' OR c.targeting->>'gender' = ?))
			AND ((c.targeting->>'age_from' IS NULL OR ? >= CAST(c.targeting->>'age_from' AS int)))
			AND ((c.targeting->>'age_to' IS NULL OR ? <= CAST(c.targeting->>'age_to' AS int)))
			AND ((c.targeting->>'location' IS NULL OR c.targeting->>'location' = ?))
		`, day, day, client.Gender, client.Age, client.Age, client.Location).
		Order("score DESC").
		Limit(1)
	if moderation {
		log.Println("moderation enabled")
		req = req.Where("moderated = ?", true)
	}
	res := req.Scan(&campaignsWithScores)
	log.Println(campaignsWithScores.AdID, campaignsWithScores.Score)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	log.Printf("find with score %f\n", campaignsWithScores.Score)
	// херачим статистику ->
	dryRun := false
	if !dryRun {
		s.db.Create(&models.Event{
			ClientID:   client.ClientID,
			CampaignID: uuid.MustParse(campaignsWithScores.AdID),
			Type:       "view",
			Cost:       campaignsWithScores.CostPerImpression,
			Day:        day,
		})
	}
	return &campaignsWithScores, nil
}

func (s *AdService) ClickAd(campaignID uuid.UUID, clientID uuid.UUID, day uint) error {
	var campaign models.Campaign
	if err := s.db.First(&campaign, "campaign_id = ?", campaignID).Error; err != nil {
		return err
	}
	var userView int64
	s.db.Model(&models.Event{}).Where("campaign_id = ?", campaignID).Where("client_id = ?", clientID).Where("type = 'view'").Count(&userView)
	var userClicks int64
	s.db.Model(&models.Event{}).Where("campaign_id = ?", campaignID).Where("client_id = ?", clientID).Where("type = 'click'").Count(&userClicks)
	if userView == 0 || userClicks != 0 {
		return gorm.ErrInvalidTransaction
	}
	var client models.Client
	if err := s.db.First(&client, "client_id = ?", clientID).Error; err != nil {
		return err
	}
	s.db.Create(&models.Event{
		ClientID:   clientID,
		CampaignID: campaign.CampaignID,
		Type:       "click",
		Cost:       *campaign.CostPerClick,
		Day:        day,
	})
	return nil
}
