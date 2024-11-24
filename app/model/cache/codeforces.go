package cache

import (
	"context"
	"time"
)

var (
	ctx = context.Background()
)

func keyCodeforcesUser(handle string) string {
	return "codeforces_user:" + handle
}

func GetCodeforcesUser(handle string) (string, error) {
	return rdb.Get(ctx, keyCodeforcesUser(handle)).Result()
}

func SetCodeforcesUser(handle string, value []byte, exp time.Duration) error {
	return rdb.Set(ctx, keyCodeforcesUser(handle), value, exp).Err()
}

func DeleteCodeforcesUser(handle string) (err error) {
	return rdb.Del(ctx, keyCodeforcesUser(handle)).Err()
}
