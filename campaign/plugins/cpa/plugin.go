package cpa

import (
	"fmt"
	"sort"

	"flinters/campaign/models"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/plugins/aggregator"
)

// Plugin computes Cost Per Action from aggregates and provides the top 10 results.
type Plugin struct{}

func (p *Plugin) Name() string { return "CPAPlugin" }

func (p *Plugin) Requires() []core.DataKey { return []core.DataKey{aggregator.KeyAggregates.Key} }

func (p *Plugin) Provides() []core.DataKey { return []core.DataKey{models.KeyTopCPA.Key} }

func (p *Plugin) Cleanup(ctx core.Context) error { return nil }

func (p *Plugin) Execute(ctx core.Context) error {
	aggregates, err := aggregator.KeyAggregates.Get(ctx)
	if err != nil {
		return fmt.Errorf("aggregates error: %w", err)
	}
	var validCPAs []models.CampaignResult

	for campaignID, groupCtx := range aggregates {
		spend, _ := core.NewTypedKey[float64]("spend").Get(groupCtx)
		conversions, _ := core.NewTypedKey[float64]("conversions").Get(groupCtx)
		impressions, _ := core.NewTypedKey[float64]("impressions").Get(groupCtx)
		clicks, _ := core.NewTypedKey[float64]("clicks").Get(groupCtx)

		if conversions > 0 {
			// CPA = SUM(spend) / NULLIF(SUM(conversions), 0)
			v := spend / conversions
			var ctr float64
			if impressions > 0 {
				ctr = clicks / impressions
			}
			validCPAs = append(validCPAs, models.CampaignResult{
				CampaignID:  campaignID,
				Impressions: impressions,
				Clicks:      clicks,
				Spend:       spend,
				Conversions: conversions,
				CTR:         ctr,
				CPA:         &v,
			})
		}
	}

	sort.Slice(validCPAs, func(i, j int) bool {
		if *validCPAs[i].CPA == *validCPAs[j].CPA {
			return validCPAs[i].CampaignID < validCPAs[j].CampaignID
		}
		return *validCPAs[i].CPA < *validCPAs[j].CPA
	})

	limit := 10
	if len(validCPAs) < limit {
		limit = len(validCPAs)
	}
	models.KeyTopCPA.Set(ctx, validCPAs[:limit])

	return nil
}
