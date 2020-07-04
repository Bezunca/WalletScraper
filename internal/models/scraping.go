package models

type ScrapingCredentials struct {
	CEI *CEI `json:"cei"`
}

type Scraping struct {
	ID                  []uint8             `json:"id"`
	ScrapingCredentials ScrapingCredentials `json:"scraping_credentials"`
}
