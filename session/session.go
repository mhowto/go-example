package session

import (
	"context"

	"github.com/twinj/uuid"
	log "gitlab.ucloudadmin.com/wu/logrus"
	"google.golang.org/grpc/metadata"
)

func NewContext(ctx context.Context) context.Context {
	md, ok := metadata.FromContext(ctx)
	if ok {
		_, ok = md["session-id"]
	}
	if !ok {
		return metadata.NewContext(ctx, metadata.Pairs("session-id", SessionId()))
	}
	return ctx
}

func NewContextWithSessionId(ctx context.Context, sessionId string) context.Context {
	return metadata.NewContext(ctx, metadata.Pairs("session-id", sessionId))
}

func SessionId() string {
	return uuid.NewV4().String()
}

func SessionIdFromContext(ctx context.Context) string {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ""
	}
	sessionIds, ok := md["session-id"]
	if !ok {
		return ""
	}
	return sessionIds[0]
}

func SessionLogEntry(ctx context.Context) *log.Entry {
	return log.WithField("session-id", SessionIdFromContext(ctx))
}
