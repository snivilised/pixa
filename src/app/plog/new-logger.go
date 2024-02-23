package plog

import (
	"log/slog"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/samber/lo"
	"github.com/snivilised/cobrass/src/assistant/configuration"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/extendio/xfs/utils"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func New(lc common.LoggingConfig,
	vfs storage.VirtualFS,
	scope common.ConfigScope,
	vc configuration.ViperConfig,
) *slog.Logger {
	logPath := lo.TernaryF(common.IsUsingXDG(vc),
		func() string {
			// manual XDG: ~/.local/share/app/filename.log
			//
			return utils.ResolvePath(filepath.Join(
				"~", ".local", "share",
				common.Definitions.Pixa.AppName,
				common.Definitions.Defaults.Logging.LogFilename,
			))
		},
		func() string {
			lp := lc.Path()
			if lp != "" {
				return utils.ResolvePath(lp)
			}

			dir, _ := scope.LogPath(common.Definitions.Defaults.Logging.LogFilename)

			return dir
		},
	)

	logPath, _ = utils.EnsurePathAt(
		logPath,
		common.Definitions.Defaults.Logging.LogFilename,
		int(common.Permissions.Write),
		vfs,
	)

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
