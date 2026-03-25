package core

import (
	"context"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

// KeyErrGroup manages concurrent background plugin operations and aggregates errors universally.
var KeyErrGroup = NewTypedKey[*errgroup.Group]("loader_errgroup")

// KeyContext holds the cancellable lifecycle context bound to the Pipeline's ErrGroup execution.
var KeyContext = NewTypedKey[context.Context]("loader_context")

// KeyLogger injects the pipeline-level structured logger.
var KeyLogger = NewTypedKey[*slog.Logger]("loader_logger")
