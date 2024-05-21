package logger

import (
	"context"
	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

type SentryHook struct {
	levels []logrus.Level
}

func NewSentryHook(levels []logrus.Level) *SentryHook {
	return &SentryHook{
		levels: levels,
	}
}

func (h *SentryHook) Levels() []logrus.Level {
	return h.levels
}

func (h *SentryHook) Fire(entry *logrus.Entry) error {
	sentryLevel := fromLogrusLevel(entry.Level)
	localHub := getOrCreateLocalHub(entry.Context)
	localScope := localHub.Scope()
	localScope.SetLevel(sentryLevel)
	localScope.SetExtra("caller", entry.Caller)

	if sentryLevel == sentry.LevelFatal || sentryLevel == sentry.LevelError {
		localScope.SetExtra("fields", entry.Data)
		localHub.CaptureMessage(entry.Message)
	} else {
		hints := sentry.BreadcrumbHint(entry.Data)
		localHub.AddBreadcrumb(&sentry.Breadcrumb{
			Category:  string(sentryLevel),
			Message:   entry.Message,
			Level:     sentryLevel,
			Timestamp: entry.Time,
		}, &hints)
	}
	return nil
}

func fromLogrusLevel(level logrus.Level) sentry.Level {
	switch level {
	case logrus.DebugLevel:
		return sentry.LevelDebug
	case logrus.InfoLevel:
		return sentry.LevelInfo
	case logrus.WarnLevel:
		return sentry.LevelWarning
	case logrus.ErrorLevel:
		return sentry.LevelError
	case logrus.FatalLevel:
		return sentry.LevelFatal
	case logrus.PanicLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelError
	}
}

func getOrCreateLocalHub(ctx context.Context) *sentry.Hub { // FIXME: sometimes throws panic: runtime error: invalid memory address or nil pointer dereference (when trying to get something from context)
	if sentry.HasHubOnContext(ctx) {
		return sentry.GetHubFromContext(ctx)
	} else {
		return sentry.CurrentHub().Clone()
	}
}
