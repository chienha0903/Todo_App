package redisstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/chienha0903/Todo_App/services/users/internal/domain/entity"
	"github.com/chienha0903/Todo_App/services/users/internal/domain/gateway"
)

// Key design:
//   rt:{tokenHash}   → JSON data của token, TTL = thời gian còn lại
//   urt:{userID}     → Redis Set chứa danh sách tokenHash của 1 user
//                      (dùng cho DeleteByUserID)

const (
	tokenKeyPrefix = "rt:"
	userKeyPrefix  = "urt:"
)

// tokenRecord là struct được serialize vào Redis.
// Không dùng entity.RefreshToken trực tiếp vì entity có field ID
// không có ý nghĩa trong Redis storage.
type tokenRecord struct {
	UserID    int64      `json:"user_id"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

type RefreshTokenRepo struct {
	client *redis.Client
}

func NewRefreshTokenRepo(client *redis.Client) *RefreshTokenRepo {
	return &RefreshTokenRepo{client: client}
}

// NewRefreshTokenCommandGateway bind repo vào interface — Wire dùng hàm này.
func NewRefreshTokenCommandGateway(r *RefreshTokenRepo) gateway.RefreshTokenCommandGateway {
	return r
}

// NewRefreshTokenQueryGateway bind repo vào interface — Wire dùng hàm này.
func NewRefreshTokenQueryGateway(r *RefreshTokenRepo) gateway.RefreshTokenQueryGateway {
	return r
}

func tokenKey(hash string) string  { return tokenKeyPrefix + hash }
func userKey(userID int64) string  { return fmt.Sprintf("%s%d", userKeyPrefix, userID) }

// StoreRefreshToken lưu token vào Redis với TTL tự động.
// Đồng thời ghi hash vào Set của user để DeleteByUserID có thể tìm được.
func (r *RefreshTokenRepo) StoreRefreshToken(ctx context.Context, token *entity.RefreshToken) error {
	rec := tokenRecord{
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt,
	}

	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("redis StoreRefreshToken marshal: %w", err)
	}

	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		return fmt.Errorf("redis StoreRefreshToken: token already expired")
	}

	// Pipeline gom 3 lệnh thành 1 round-trip đến Redis
	pipe := r.client.Pipeline()
	pipe.Set(ctx, tokenKey(token.TokenHash), data, ttl)
	pipe.SAdd(ctx, userKey(token.UserID), token.TokenHash)
	pipe.Expire(ctx, userKey(token.UserID), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis StoreRefreshToken pipeline: %w", err)
	}

	return nil
}

// FindByTokenHash tìm token theo hash.
// Trả về nil, nil nếu không tìm thấy (đã hết hạn hoặc không tồn tại).
func (r *RefreshTokenRepo) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	data, err := r.client.Get(ctx, tokenKey(tokenHash)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // không tìm thấy → trả nil như contract của interface
		}
		return nil, fmt.Errorf("redis FindByTokenHash: %w", err)
	}

	var rec tokenRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("redis FindByTokenHash unmarshal: %w", err)
	}

	return &entity.RefreshToken{
		UserID:    rec.UserID,
		TokenHash: tokenHash,
		ExpiresAt: rec.ExpiresAt,
		UsedAt:    rec.UsedAt,
	}, nil
}

// MarkUsed đánh dấu token đã được dùng bằng cách set field used_at.
// Dùng KeepTTL để không reset thời gian hết hạn của key.
func (r *RefreshTokenRepo) MarkUsed(ctx context.Context, tokenHash string, usedAt time.Time) error {
	key := tokenKey(tokenHash)

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil // token đã hết hạn, không cần làm gì
		}
		return fmt.Errorf("redis MarkUsed get: %w", err)
	}

	var rec tokenRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return fmt.Errorf("redis MarkUsed unmarshal: %w", err)
	}

	rec.UsedAt = &usedAt

	updated, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("redis MarkUsed marshal: %w", err)
	}

	// redis.KeepTTL = giữ nguyên TTL hiện tại, không reset về 0
	if err := r.client.Set(ctx, key, updated, redis.KeepTTL).Err(); err != nil {
		return fmt.Errorf("redis MarkUsed set: %w", err)
	}

	return nil
}

// DeleteByUserID xóa tất cả token của 1 user.
// Dùng khi user đổi mật khẩu hoặc bị xóa.
func (r *RefreshTokenRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	uKey := userKey(userID)

	// Lấy danh sách token hashes của user
	hashes, err := r.client.SMembers(ctx, uKey).Result()
	if err != nil {
		return fmt.Errorf("redis DeleteByUserID smembers: %w", err)
	}

	if len(hashes) == 0 {
		return nil
	}

	// Xóa tất cả rt:{hash} + urt:{userID} trong 1 lệnh DEL
	keys := make([]string, 0, len(hashes)+1)
	for _, h := range hashes {
		keys = append(keys, tokenKey(h))
	}
	keys = append(keys, uKey)

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis DeleteByUserID del: %w", err)
	}

	return nil
}
