package signing

import (
	"context"

	"github.com/transparency-dev/tessera"
	"github.com/transparency-dev/tessera/storage/posix"
)

type LogOperator struct {
	appender           *tessera.Appender
	reader             tessera.LogReader
	publicationAwaiter *tessera.PublicationAwaiter
	shutdown           func(ctx context.Context) error
}

func OperateLog(ctx context.Context, logPath string, cfg *tessera.AppendOptions) (*LogOperator, error) {
	driver, err := posix.New(ctx, posix.Config{
		Path: logPath,
	})
	if err != nil {
		return nil, err
	}
	lop := LogOperator{}

	lop.appender, lop.shutdown, lop.reader, err = tessera.NewAppender(
		ctx, driver, cfg)
	if err != nil {
		return nil, err
	}

	lop.publicationAwaiter = tessera.NewPublicationAwaiter(
		ctx, lop.reader.ReadCheckpoint, cfg.BatchMaxAge())

	return &lop, nil
}

func (lop *LogOperator) Shutdown(ctx context.Context) error {
	return lop.shutdown(ctx)
}
