package event

import "encoding/json"

const (
	CacheAggregateType         = "cache"
	CacheEventDeleteUserTokens = "cache.delete_user_tokens"
)

type DeleteUserTokensPayload struct {
	UserID int64 `json:"user_id"`
}

func (p *DeleteUserTokensPayload) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}
