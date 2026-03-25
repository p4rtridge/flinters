package ctr

import (
	"fmt"
	"sort"

	"flinters/campaign/models"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/plugins/aggregator"
)

// Plugin computes Click-Through Rate from aggregates and provides the top 10 results.
type Plugin struct{}

func (p *Plugin) Name() string { return "CTRPlugin" }

func (p *Plugin) Requires() []core.DataKey { return []core.DataKey{aggregator.KeyAggregates.Key} }

func (p *Plugin) Provides() []core.DataKey { return []core.DataKey{models.KeyTopCTR.Key} }

func (p *Plugin) Cleanup(ctx core.Context) error { return nil }

func (p *Plugin) Execute(ctx core.Context) error {
	aggregates, err := aggregator.KeyAggregates.Get(ctx)
	if err != nil {
		return fmt.Errorf("aggregates error: %w", err)
	}

	var results []models.CampaignResult

	for campaignID, groupCtx := range aggregates {
		clicks, _ := core.NewTypedKey[float64]("clicks").Get(groupCtx)
		impressions, _ := core.NewTypedKey[float64]("impressions").Get(groupCtx)
		spend, _ := core.NewTypedKey[float64]("spend").Get(groupCtx)
		conversions, _ := core.NewTypedKey[float64]("conversions").Get(groupCtx)

		var ctr float64
		if impressions > 0 {
			ctr = clicks / impressions
		}

		// CPA = SUM(spend) / NULLIF(SUM(conversions), 0) → nil when conversions = 0
		var cpa *float64
		if conversions > 0 {
			v := spend / conversions
			cpa = &v
		}

		results = append(results, models.CampaignResult{
			CampaignID:  campaignID,
			Impressions: impressions,
			Clicks:      clicks,
			Spend:       spend,
			Conversions: conversions,
			CTR:         ctr,
			CPA:         cpa,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].CTR == results[j].CTR {
			return results[i].CampaignID < results[j].CampaignID
		}
		return results[i].CTR > results[j].CTR
	})

	limit := 10
	if len(results) < limit {
		limit = len(results)
	}
	models.KeyTopCTR.Set(ctx, results[:limit])

	return nil
}
