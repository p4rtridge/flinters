package writer

import (
	"fmt"

	"flinters/campaign/models"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/plugins/csvwriter"
)

type CTRWriterPlugin struct{}

func (p *CTRWriterPlugin) Name() string             { return "CTRWriterPlugin" }
func (p *CTRWriterPlugin) Requires() []core.DataKey { return []core.DataKey{models.KeyTopCTR.Key} }
func (p *CTRWriterPlugin) Provides() []core.DataKey {
	return []core.DataKey{csvwriter.KeyExport.Key}
}

func (p *CTRWriterPlugin) Cleanup(ctx core.Context) error { return nil }

func (p *CTRWriterPlugin) Execute(ctx core.Context) error {
	data, err := models.KeyTopCTR.Get(ctx)
	if err != nil {
		return fmt.Errorf("ctr data missing: %w", err)
	}

	stream := func(yield func([]string) bool) {
		for _, stat := range data {
			cpaStr := "null"
			if stat.CPA != nil {
				cpaStr = fmt.Sprintf("%.4f", *stat.CPA)
			}
			row := []string{
				stat.CampaignID,
				fmt.Sprintf("%.0f", stat.Impressions),
				fmt.Sprintf("%.0f", stat.Clicks),
				fmt.Sprintf("%.4f", stat.Spend),
				fmt.Sprintf("%.0f", stat.Conversions),
				fmt.Sprintf("%.6f", stat.CTR),
				cpaStr,
			}
			if !yield(row) {
				break
			}
		}
	}

	export := csvwriter.Export{
		Filename: "top_10_ctr.csv",
		Header:   []string{"campaign_id", "total_impressions", "total_clicks", "total_spend", "total_conversions", "CTR", "CPA"},
		Stream:   stream,
	}

	csvwriter.KeyExport.Set(ctx, export)

	return nil
}

type CPAWriterPlugin struct{}

func (p *CPAWriterPlugin) Name() string { return "CPAWriterPlugin" }

func (p *CPAWriterPlugin) Requires() []core.DataKey { return []core.DataKey{models.KeyTopCPA.Key} }

func (p *CPAWriterPlugin) Provides() []core.DataKey {
	return []core.DataKey{csvwriter.KeyExport.Key}
}

func (p *CPAWriterPlugin) Cleanup(ctx core.Context) error { return nil }

func (p *CPAWriterPlugin) Execute(ctx core.Context) error {
	data, err := models.KeyTopCPA.Get(ctx)
	if err != nil {
		return fmt.Errorf("cpa data missing: %w", err)
	}

	stream := func(yield func([]string) bool) {
		for _, stat := range data {
			cpaStr := "null"
			if stat.CPA != nil {
				cpaStr = fmt.Sprintf("%.4f", *stat.CPA)
			}
			row := []string{
				stat.CampaignID,
				fmt.Sprintf("%.0f", stat.Impressions),
				fmt.Sprintf("%.0f", stat.Clicks),
				fmt.Sprintf("%.4f", stat.Spend),
				fmt.Sprintf("%.0f", stat.Conversions),
				fmt.Sprintf("%.6f", stat.CTR),
				cpaStr,
			}
			if !yield(row) {
				break
			}
		}
	}

	export := csvwriter.Export{
		Filename: "top_10_cpa.csv",
		Header:   []string{"campaign_id", "total_impressions", "total_clicks", "total_spend", "total_conversions", "CTR", "CPA"},
		Stream:   stream,
	}

	csvwriter.KeyExport.Set(ctx, export)

	return nil
}
