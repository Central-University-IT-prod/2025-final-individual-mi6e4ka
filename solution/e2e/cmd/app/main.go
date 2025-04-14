package main

import (
	"log"
	"math/rand"
	"time"

	"git.mi6e4ka.dev/prod-2025-e2e/pkg/http_req"
	"git.mi6e4ka.dev/prod-2025-e2e/pkg/structs"
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
)

func main() {
	gofakeit.Seed(time.Now().UnixNano())
	api := http_req.HTTPReq{BaseURL: "http://localhost:8080"}
	var tmpStats map[string]*map[string]*int
	CorruptedCreateClients(api)
	advertisers := BulkCreateAdvertisers(api, 10)
	clients := BulkCreateClients(api, 100)

	for _, adv := range advertisers {
		var advBody structs.Advertiser
		status, _ := api.GET("/advertisers/"+adv.AdvertiserID.String(), &advBody)
		if status != 200 {
			log.Fatalf("expected advertiser %s to exist", adv.AdvertiserID.String())
		}
		if advBody.Name != adv.Name {
			log.Fatalf("expected to match names")
		}
	}
	log.Println("all advertisers exists")
	for _, cli := range clients {
		var cliBody structs.Client
		status, _ := api.GET("/clients/"+cli.ClientID.String(), &cliBody)
		if status != 200 {
			log.Fatalf("expected client %s to exist", cli.ClientID.String())
		}
		if cliBody.Login != cli.Login {
			log.Fatalf("expected to match logins")
		}
	}
	log.Println("all clients exists")
	SetRandomMLScores(api, clients, advertisers)
	campaigns := CreateRandomCampaigns(api, advertisers)
	_ = campaigns
	for day := 0; day <= 40; day++ {
		status, _ := api.POST("/time/advance", structs.TimeSetBody{CurrentDate: day}, nil)
		if status != 200 {
			log.Fatalf("failed to set day %d", status)
		}
		for _, client := range clients {
			var ad structs.AdUser
			status, _ := api.GET("/ads?client_id="+client.ClientID.String(), &ad)
			if status != 200 && status != 404 {
				log.Fatalf("excepted status 200 or 404, received %d", status)
			}
			if status == 200 {
				log.Printf("received ad %s", ad.AdID)
				if tmpStats[ad.AdID] != nil {
					if (*tmpStats[ad.AdID])["impressions"] != nil {
						(*tmpStats[ad.AdID])["impressions"] = ptrInt(*(*tmpStats[ad.AdID])["impressions"] + 1)
					} else {
						(*tmpStats[ad.AdID])["impressions"] = ptrInt(1)
					}
				}
				if rand.Intn(101) > 85 {
					log.Printf("click ad %s", ad.AdID)

					status, _ := api.POST("/ads/"+ad.AdID+"/click", structs.QueryClient{ClientID: client.ClientID.String()}, nil)
					if status != 204 {
						log.Fatalf("excepted status 204, received %d", status)
					}
					if tmpStats[ad.AdID] != nil {
						if (*tmpStats[ad.AdID])["clicks"] != nil {
							(*tmpStats[ad.AdID])["clicks"] = ptrInt(*(*tmpStats[ad.AdID])["clicks"] + 1)
						} else {
							(*tmpStats[ad.AdID])["clicks"] = ptrInt(1)
						}
					}
				}
			} else {
				log.Printf("received 404")
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

}
func CreateRandomCampaigns(api http_req.HTTPReq, advertisers []structs.Advertiser) []structs.Campaign {
	var campaigns []structs.Campaign
	for _, advertiser := range advertisers {
		iter := rand.Intn(2) + 1
		for i := 0; i <= iter; i++ {
			startDate := rand.Intn(30)
			targeting := structs.Targeting{}
			targeting.Gender = []*string{ptrStr("MALE"), ptrStr("FEMALE"), nil, nil, nil, nil}[rand.Intn(6)]
			if rand.Intn(100) > 80 {
				targeting.AgeFrom = ptrInt(rand.Intn(50) + 14)
			}
			if rand.Intn(100) > 80 {
				if targeting.AgeFrom != nil {
					targeting.AgeTo = ptrInt(rand.Intn(25) + *targeting.AgeFrom)
				} else {
					targeting.AgeTo = ptrInt(rand.Intn(50) + 14)
				}
			}
			if rand.Intn(101) > 90 {
				targeting.Location = ptrStr(gofakeit.Country())
			}
			clicksLimit := rand.Intn(200) + 50
			campaign := structs.Campaign{
				AdvertiserID:      advertiser.AdvertiserID,
				ImpressionsLimit:  clicksLimit * 2,
				ClicksLimit:       clicksLimit,
				CostPerImpression: float64(rand.Intn(100)+10) / 10,
				CostPerClick:      float64(rand.Intn(300)+100) / 10,
				AdTitle:           gofakeit.HackerPhrase(),
				AdText:            gofakeit.Quote(),
				StartDate:         startDate,
				EndDate:           startDate + rand.Intn(10),
				Targeting:         targeting,
			}
			campaigns = append(campaigns, campaign)
			status, _ := api.POST("/advertisers/"+advertiser.AdvertiserID.String()+"/campaigns", campaign, nil)
			if status != 201 {
				log.Fatalf("excepted status 201, received %d", status)
			}
		}
	}
	log.Printf("successful created %d campaigns", len(campaigns))
	return campaigns
}
func SetRandomMLScores(api http_req.HTTPReq, clients []structs.Client, advertisers []structs.Advertiser) {
	for _, client := range clients {
		for _, advertiser := range advertisers {
			status, _ := api.POST("/ml-scores", structs.MLScore{
				AdvertiserID: advertiser.AdvertiserID,
				ClientID:     client.ClientID,
				Score:        uint(rand.Intn(10000)),
			}, nil)
			if status != 200 {
				log.Fatalf("excepted status 200, received %d", status)
			}
		}
	}
	log.Printf("successful created %d ml scores", len(clients)*len(advertisers))
}

func BulkCreateAdvertisers(api http_req.HTTPReq, count int) []structs.Advertiser {
	advertisers := []structs.Advertiser{}
	for i := 0; i < count; i++ {
		advertisers = append(advertisers, structs.Advertiser{
			AdvertiserID: uuid.New(),
			Name:         gofakeit.Company(),
		})
	}
	// log.Println(advertisers)

	status, err := api.POST("/advertisers/bulk", &advertisers, nil)
	if err != nil {
		panic(err)
	}
	if status != 201 {
		log.Fatalf("expected status code 201, received %d", status)
	}
	log.Printf("successful created %d advertisers", count)
	return advertisers
}
func BulkCreateClients(api http_req.HTTPReq, count int) []structs.Client {
	clients := []structs.Client{}
	for i := 0; i < count; i++ {
		clients = append(clients, structs.Client{
			ClientID: uuid.New(),
			Login:    gofakeit.Username(),
			Age:      rand.Intn(65-14+1) + 14,
			Location: gofakeit.Country(),
			Gender:   []string{"MALE", "FEMALE"}[rand.Intn(2)],
		})
	}
	// log.Println(clients)

	status, err := api.POST("/clients/bulk", &clients, nil)
	if err != nil {
		panic(err)
	}
	if status != 201 {
		log.Fatalf("expected status code 201, received %d", status)
	}
	log.Printf("successful created %d clients", count)
	return clients
}
func CorruptedCreateClients(api http_req.HTTPReq) {
	corruptedClients := []structs.Client{
		{
			ClientID: uuid.New(),
			Login:    gofakeit.Username(),
			Age:      rand.Intn(65-14+1) + 14,
			Location: gofakeit.Address().Address,
			Gender:   "HELICOPTER",
		},
		{
			ClientID: uuid.New(),
			Login:    gofakeit.Username(),
			Age:      -10,
			Location: gofakeit.Address().Address,
			Gender:   []string{"MALE", "FEMALE"}[rand.Intn(2)],
		},
	}
	for _, client := range corruptedClients {
		status, err := api.POST("/clients/bulk", &[]structs.Client{client}, nil)
		if err != nil {
			panic(err)
		}
		if status == 201 {
			log.Fatalf("expected status code 400, received %d", status)
		}
	}
	log.Printf("test corrupted clients")
}

func ptrStr(str string) *string {
	return &str
}
func ptrInt(num int) *int {
	return &num
}
