package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/plugins/aggregator"
	"github.com/p4rtridge/p4rse_tan/plugins/csvreader"
	"github.com/p4rtridge/p4rse_tan/plugins/filereader"
)

var (
	rowsFlag = flag.Int("rows", 100_000, "Number of rows to generate and process")
	modeFlag = flag.String("mode", "", "Mode: gen | run_p4rse")
)

const (
	fileName = "benchmark_data.csv"
)

var regions = []string{"US", "EU", "VN", "JP", "CN", "DE", "FR", "UK", "CA", "AU"}

func main() {
	flag.Parse()

	switch *modeFlag {
	case "gen":
		generateData(*rowsFlag)
	case "run_p4rse":
		runP4rseTan()
	default:
		fmt.Println("Unknown mode. Use -mode=gen or -mode=run_p4rse")
		os.Exit(1)
	}
}

func generateData(rows int) {
	fmt.Printf("Generating %d rows to %s...\n", rows, fileName)
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("region,amount\n")

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < rows; i++ {
		region := regions[rng.Intn(len(regions))]
		amount := rng.Float64() * 1000.0
		fmt.Fprintf(w, "%s,%.2f\n", region, amount)
	}
	w.Flush()
	fmt.Println("Done.")
}

func runP4rseTan() {
	ctx := make(core.Context)
	execCtx := context.Background()

	// FileReader -> CSVReader -> Aggregator -> Printer
	// Using wrapper style as per current implementation of csvreader
	fileReader := &filereader.Plugin{FilePath: fileName}
	csvReader := &csvreader.Plugin{Input: fileReader}

	aggregatorPlugin := &aggregator.Plugin{
		GroupKey: "region",
		SumKeys:  []string{"amount"},
	}
	printer := &PrinterPlugin{}

	plugins := []core.Plugin{
		csvReader,        // Executes FileReader internally
		aggregatorPlugin, // Consumes stream from csvReader
		printer,          // Consumes aggregates
	}

	pipeline, err := core.BuildPipeline(ctx, plugins)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	if err := pipeline.Execute(execCtx, ctx); err != nil {
		panic(err)
	}
	fmt.Printf("p4rse_tan duration: %v\n", time.Since(start))
}

type PrinterPlugin struct{}

func (p *PrinterPlugin) Name() string { return "PrinterPlugin" }
func (p *PrinterPlugin) Requires() []core.DataKey {
	return []core.DataKey{aggregator.KeyAggregates.Key}
}
func (p *PrinterPlugin) Provides() []core.DataKey { return nil }
func (p *PrinterPlugin) Execute(ctx core.Context) error {
	aggs, err := aggregator.KeyAggregates.Get(ctx)
	if err != nil {
		return err
	}
	// Print a few results to verify correctness without flooding stdout
	count := 0
	fmt.Printf("Total groups: %d\n", len(aggs))
	for k, v := range aggs {
		if count < 5 {
			// Extract sum
			sumKey := core.NewTypedKey[float64]("amount")
			sum, _ := sumKey.Get(v)
			fmt.Printf("Region: %s, Sum: %.2f\n", k, sum)
		}
		count++
	}
	return nil
}
func (p *PrinterPlugin) Cleanup(ctx core.Context) error { return nil }
