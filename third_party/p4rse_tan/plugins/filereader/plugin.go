package filereader

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/logger"
)

type Plugin struct {
	FilePath string
}

func (p *Plugin) Name() string { return "FileReaderPlugin" }

func (p *Plugin) Requires() []core.DataKey {
	return nil
}

func (p *Plugin) Provides() []core.DataKey {
	return []core.DataKey{core.KeyImport.Key}
}

func (p *Plugin) Cleanup(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	importData, err := core.KeyImport.Get(ctx)
	if err == nil && importData.Reader != nil {
		log.Debug("[DEBUG] closing import reader", slog.String("file", importData.Name))
		return importData.Reader.Close()
	}

	return nil
}

func (p *Plugin) Execute(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	log.Debug("[DEBUG] opening file", slog.String("path", p.FilePath))

	f, err := os.Open(p.FilePath)
	if err != nil {
		return fileReadErr(fmt.Sprintf("open %s", p.FilePath), err)
	}

	log.Debug("[DEBUG] file opened", slog.String("path", p.FilePath))

	core.KeyImport.Set(ctx, core.Import{
		Name: p.FilePath,
		Reader: struct {
			*bufio.Reader
			io.Closer
		}{
			Reader: bufio.NewReaderSize(f, 4*1024*1024), // Explicitly enforcing 4MB bounds for exceptionally long JSON/CSV rows.
			Closer: f,
		},
	})

	return nil
}

func fileReadErr(reason string, cause error) error {
	if cause == nil {
		return fmt.Errorf("filereader %s", reason)
	}
	return fmt.Errorf("filereader %s: %w", reason, cause)
}
