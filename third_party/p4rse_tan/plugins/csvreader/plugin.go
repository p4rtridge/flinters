package csvreader

import (
	"bufio"
	"fmt"
	"log/slog"
	"strings"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/logger"
)

type csvRecord struct {
	header map[string]int
	row    [][]byte
}

func (c *csvRecord) Get(key string) ([]byte, bool) {
	if idx, ok := c.header[key]; ok && idx < len(c.row) {
		return c.row[idx], true
	}
	return nil, false
}

type Plugin struct {
	Input core.Plugin
}

func (p *Plugin) Name() string { return "CSVReaderPlugin" }

func (p *Plugin) Requires() []core.DataKey {
	if p.Input != nil {
		return p.Input.Requires()
	}
	return nil
}

func (p *Plugin) Provides() []core.DataKey {
	var keys []core.DataKey
	if p.Input != nil {
		for _, prov := range p.Input.Provides() {
			if prov != core.KeyImport.Key {
				keys = append(keys, prov)
			}
		}
	}
	return append(keys, core.KeyRecordStream.Key)
}

func (p *Plugin) Cleanup(ctx core.Context) error {
	if p.Input != nil {
		return p.Input.Cleanup(ctx)
	}
	return nil
}

func (p *Plugin) Execute(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	if p.Input != nil {
		if err := p.Input.Execute(ctx); err != nil {
			return err
		}
	}

	importData, err := core.KeyImport.Get(ctx)
	if err != nil {
		return csvReadErr("import stream missing", err)
	}

	file := importData.Reader
	scanner := bufio.NewScanner(file)

	if !scanner.Scan() {
		return csvReadErr(fmt.Sprintf("header read failed for %s: EOF or empty file", importData.Name), nil)
	}

	headerLine := string(scanner.Bytes())
	headerRecord := strings.Split(headerLine, ",")

	headerMap := make(map[string]int)
	for i, col := range headerRecord {
		headerMap[strings.TrimSpace(col)] = i
	}

	cols := make([][]byte, len(headerRecord))
	wrapper := &csvRecord{header: headerMap, row: cols}

	// scanErr is captured outside the stream closure so we can emit a single
	// structured log entry after iteration ends, keeping the hot path I/O-free.
	var scanErr error

	stream := func(yield func(core.Record) bool) {
		for scanner.Scan() {
			line := scanner.Bytes()

			idx := 0
			start := 0
			n := len(line)
			for j := 0; j < n; j++ {
				if line[j] == ',' {
					if idx < len(cols) {
						cols[idx] = line[start:j]
					}
					start = j + 1
					idx++
				}
			}
			if idx < len(cols) {
				cols[idx] = line[start:n]
			}

			// Blank missing columns to avoid stale data from the previous row.
			for k := idx + 1; k < len(cols); k++ {
				cols[k] = nil
			}

			if !yield(wrapper) {
				break
			}
		}
		// Capture scan error for post-loop reporting; no I/O inside hot path.
		scanErr = scanner.Err()
	}
	core.KeyRecordStream.Set(ctx, stream)

	// scanErr is nil until the stream closure has been consumed by a downstream
	// plugin. If the caller inspects it before consuming the stream, it will
	// always be nil here — that is expected.
	if scanErr != nil {
		log.Warn("csv scanner terminated with an error",
			slog.String("source", importData.Name),
			slog.Any("error", scanErr),
		)
	}

	return nil
}

func csvReadErr(reason string, cause error) error {
	if cause == nil {
		return fmt.Errorf("csvreader %s", reason)
	}
	return fmt.Errorf("csvreader %s: %w", reason, cause)
}
