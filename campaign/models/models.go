package models

import "github.com/p4rtridge/p4rse_tan/core"

// CampaignResult holds metrics for a single campaign.
type CampaignResult struct {
	CampaignID  string
	Impressions float64
	Clicks      float64
	Spend       float64
	Conversions float64
	CTR         float64
	CPA         *float64 // nil when conversions = 0 (SQL NULLIF semantics)
}

var KeyTopCTR = core.NewTypedKey[[]CampaignResult]("top_ctr_results")
var KeyTopCPA = core.NewTypedKey[[]CampaignResult]("top_cpa_results")
