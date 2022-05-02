package blog

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

// used for logging to a logfile
var logger zerolog.Logger
var logfile *os.File

func initLogging() {
	var err error
	logfile, err = os.OpenFile(logfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("failed when open log file")
	}

	level, err := zerolog.ParseLevel(loglevel)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse log level")
	}

	zerolog.SetGlobalLevel(level)
	logger = zerolog.New(logfile).With().Timestamp().Logger()
}

func closeLogFile() {
	if err := logfile.Close(); err != nil {
		log.Fatal().Err(err).Msg("failed when close log file")
	}
}

func Debug(s string) {
	logger.Debug().Msg(s)
}

func Info(s string) {
	logger.Info().Msg(s)
}

func Warn(s string) {
	logger.Warn().Msg(s)
}
