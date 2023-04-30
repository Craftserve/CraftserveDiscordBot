package logger

import (
	"context"
	"github.com/sirupsen/logrus"
	"io"
	"os"
)

var Logger = logrus.New()

const loggerCtxKey string = "logger"

func ConfigureLogger() {
	logFile, err := os.Create("./logs.log")
	if err != nil {
		logrus.Fatal("Could not create log file")
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	Logger.SetOutput(mw)

	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true

	Logger.SetLevel(logrus.DebugLevel)
	Logger.SetFormatter(Formatter)
}

func GetLoggerFromContext(ctx context.Context) MyLogger {
	logger, ok := ctx.Value(loggerCtxKey).(*logrus.Entry)
	if !ok {
		return MyLogger{Logger.WithContext(ctx)}
	}

	return MyLogger{logger.WithContext(ctx)}
}

func ContextWithLogger(ctx context.Context, logger MyLogger) context.Context {
	return context.WithValue(ctx, loggerCtxKey, logger)
}

type MyLogger struct {
	*logrus.Entry
}

//func (s MyLogger) Info(args ...interface{}) {
//	s.Entry.Log(logrus.InfoLevel, fmt.Sprint(args...))
//}
//
//func (s MyLogger) Error(args ...interface{}) {
//	s.Entry.Log(logrus.ErrorLevel, fmt.Sprint(args...))
//}

func (s MyLogger) WithError(err error) MyLogger {
	s.Entry = s.Entry.WithField("err", err.Error())

	return s
}

func (s MyLogger) WithGuild(guildId string) MyLogger {
	s.Entry = s.Entry.WithField("guild", guildId)

	return s
}

func (s MyLogger) WithUser(userId string) MyLogger {
	s.Entry = s.Entry.WithField("user", userId)

	return s
}

func (s MyLogger) WithMessage(messageId string) MyLogger {
	s.Entry = s.Entry.WithField("message", messageId)

	return s
}

func (s MyLogger) WithCommand(commandName string) MyLogger {
	s.Entry = s.Entry.WithField("command", commandName)

	return s
}

func (s MyLogger) WithSubcommand(subcommandName string) MyLogger {
	s.Entry = s.Entry.WithField("subcommand", subcommandName)

	return s
}
