package bootstrap

import "time"

var (
	DBMaxPool     = 25
	DBMaxIdle     = 10
	DBMaxLifeTime = time.Minute * 5 // 5 minutes
)
