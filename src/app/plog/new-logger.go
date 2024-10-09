package plog

import (
	"log/slog"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/traverse/lfs"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func New(lc common.LoggingConfig,
	tfs lfs.TraverseFS,
	scope common.ConfigScope,
	vc configuration.ViperConfig,
) *slog.Logger {
	logPath := lo.TernaryF(common.IsUsingXDG(vc),
		func() string {
			// manual XDG: ~/.local/share/app/filename.log
			//
			return lfs.ResolvePath(filepath.Join(
				"~", ".local", "share",
				common.Definitions.Pixa.AppName,
				common.Definitions.Defaults.Logging.LogFilename,
			))
		},
		func() string {
			lp := lc.Path()
			if lp != "" {
				return lfs.ResolvePath(lp)
			}

			dir, _ := scope.LogPath(common.Definitions.Defaults.Logging.LogFilename)

			return dir
		},
	)

	logPath, _ = lfs.EnsurePathAt(
		logPath,
		common.Definitions.Defaults.Logging.LogFilename,
		common.Permissions.Write, // TODO: check this is correct (file 666, or dir 777?)
		tfs,
	)

	sync := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    int(lc.MaxSizeInMb()),    //nolint:gosec // ok
		MaxBackups: int(lc.MaxNoOfBackups()), //nolint:gosec // ok
		MaxAge:     int(lc.MaxAgeInDays()),   //nolint:gosec // ok
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
