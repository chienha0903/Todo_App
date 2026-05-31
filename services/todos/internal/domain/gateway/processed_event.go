package gateway

import "context"

type ProcessedEventGateway interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	Insert(ctx context.Context, eventID string) error
}
