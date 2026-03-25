package filewriter

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/p4rtridge/p4rse_tan/core"
	"github.com/p4rtridge/p4rse_tan/logger"
)

// Plugin executes atomic OS binary disk buffering resolving universally encoded bytes
type Plugin struct {
	OutputDir string
	f         *os.File
	reader    io.ReadCloser
}

func (p *Plugin) Name() string { return "FileWriterPlugin" }

func (p *Plugin) Requires() []core.DataKey {
	return []core.DataKey{core.KeyExport.Key}
}

func (p *Plugin) Provides() []core.DataKey {
	return nil
}

func (p *Plugin) Cleanup(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	if p.f != nil {
		log.Debug("[DEBUG] closing output file", slog.String("path", p.f.Name()))
		p.f.Close()
	}

	if p.reader != nil {
		log.Debug("[DEBUG] closing export reader")
		p.reader.Close()
	}

	return nil
}

func (p *Plugin) Execute(ctx core.Context) error {
	log := logger.FromContext(ctx).With(slog.String("plugin", p.Name()))

	log.Debug("[DEBUG] creating output directory", slog.String("dir", p.OutputDir))
	if err := os.MkdirAll(p.OutputDir, os.ModePerm); err != nil {
		return fileWriteErr("create output directory", err)
	}

	export, err := core.KeyExport.Get(ctx)
	if err != nil {
		return fileWriteErr("export stream missing", err)
	}

	outPath := filepath.Join(p.OutputDir, export.Name)
	log.Debug("[DEBUG] creating output file", slog.String("path", outPath))

	f, err := os.Create(outPath)
	if err != nil {
		return fileWriteErr(fmt.Sprintf("create %s", outPath), err)
	}
	p.f = f
	p.reader = export.Reader

	log.Debug("[DEBUG] writing export to file", slog.String("path", outPath))

	eg, _ := core.KeyErrGroup.Get(ctx)
	if eg != nil {
		eg.Go(func() error {
			_, err := io.Copy(f, export.Reader)
			if err != nil {
				return fileWriteErr("copy failed", err)
			}
			log.Debug("[DEBUG] export written successfully", slog.String("path", outPath))
			return nil
		})
	} else {
		return fileWriteErr("errgroup missing from context", nil)
	}

	return nil
}

func fileWriteErr(reason string, cause error) error {
	if cause == nil {
		return fmt.Errorf("filewriter %s", reason)
	}
	return fmt.Errorf("filewriter %s: %w", reason, cause)
}
