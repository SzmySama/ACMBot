package cache

import (
	"context"
	"strings"
	"time"
)

func keyAtcoderUser(handle string) string {
	return "codeforces_user:" + strings.ToLower(handle)
}

func GetAtcoderUser(handle string) (string, error) {
	return rdb.Get(context.Background(), keyAtcoderUser(handle)).Result()
}

func SetAtcoderUser(handle string, value []byte, exp time.Duration) error {
	return rdb.Set(context.Background(), keyAtcoderUser(handle), value, exp).Err()
}

func DeleteAtcoderUser(handle string) (err error) {
	return rdb.Del(context.Background(), keyAtcoderUser(handle)).Err()
}
