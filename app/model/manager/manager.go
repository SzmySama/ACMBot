package manager

import (
	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

var (
	rdb *redis.Client
	mdb *gorm.DB
)

func init() {
	mdb = db.GetDBConnection()
	rdb = cache.GetRedisClient()
}
