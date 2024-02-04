package plog

import (
	"log/slog"

	"github.com/natefinch/lumberjack"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLogFilename = "pixa.log"
	perm               = 0o766
)

func New(lc common.LoggingConfig, vfs storage.VirtualFS) *slog.Logger {
	noc := slog.New(zapslog.NewHandler(
		zapcore.NewNopCore(), nil),
	)

	logPath := lc.Path()

	if logPath == "" {
		return noc
	}

	logPath = utils.ResolvePath(logPath)
	logPath, _ = utils.EnsurePathAt(logPath, defaultLogFilename, perm, vfs)

	sync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    int(lc.MaxSizeInMb()),
		MaxBackups: int(lc.MaxNoOfBackups()),
		MaxAge:     int(lc.MaxAgeInDays()),
	})
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout(lc.TimeFormat())
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		sync,
		level(lc.Level()),
	)

	return slog.New(zapslog.NewHandler(core, nil))
}

func level(raw string) zapcore.LevelEnabler {
	if l, err := zapcore.ParseLevel(raw); err == nil {
		return l
	}

	return zapcore.InfoLevel
}
