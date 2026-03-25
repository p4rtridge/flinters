package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/logger"
	"github.com/p4rtridge/p4rse_tan/plugins/aggregator"
	"github.com/p4rtridge/p4rse_tan/plugins/csvreader"
	"github.com/p4rtridge/p4rse_tan/plugins/csvwriter"
	"github.com/p4rtridge/p4rse_tan/plugins/filereader"
	"github.com/p4rtridge/p4rse_tan/plugins/filewriter"

	"flinters/campaign/plugins/cpa"
	"flinters/campaign/plugins/ctr"
	"flinters/campaign/plugins/writer"
)

func main() {
	var inputPath string
	var outputPath string

	flag.StringVar(&inputPath, "input", "data/ad_data.csv", "Path to input CSV file")
	flag.StringVar(&outputPath, "output", "results/", "Path to output directory")
	flag.Parse()

	execCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// 1. Initialize Global Context
	ctx := make(core.Context)
	logger.Key.Set(ctx, logger.NewDefault())

	log.Printf("Starting data-driven aggregator pipeline. Input: %s. Output: %s", inputPath, outputPath)

	// 2. Define System Plugins
	systemPlugins := []core.Plugin{
		&csvreader.Plugin{
			Input: &filereader.Plugin{
				FilePath: inputPath,
			},
		},
		&aggregator.Plugin{
			GroupKey: "campaign_id",
			SumKeys:  []string{"impressions", "clicks", "spend", "conversions"},
		},
		&ctr.Plugin{},
		&cpa.Plugin{},
		&csvwriter.Plugin{
			Input:  &writer.CTRWriterPlugin{},
			Output: &filewriter.Plugin{OutputDir: outputPath},
		},
		&csvwriter.Plugin{
			Input:  &writer.CPAWriterPlugin{},
			Output: &filewriter.Plugin{OutputDir: outputPath},
		},
	}

	// 3. Build Pipeline DAG
	pipeline, err := core.BuildPipeline(ctx, systemPlugins)
	if err != nil {
		log.Fatalf("Failed to build plugin pipeline: %v", err)
	}

	log.Printf("Pipeline successfully mapped with %d plugins.", len(pipeline.Plugins()))
	for i, p := range pipeline.Plugins() {
		log.Printf("  Step %d: %s", i+1, p.Name())
	}

	// 4. Execute Global Pipeline
	if err := pipeline.Execute(execCtx, ctx); err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	log.Printf("Done! Pipeline finished successfully. Check %s directory.", outputPath)
}
