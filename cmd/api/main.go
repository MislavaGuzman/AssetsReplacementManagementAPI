package main

import (
	"expvar"
	"runtime"
	"strconv"

	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/db"
	"github.com/MislavaGuzman/AssetsReplacementManagementAPI/internal/env"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

const version = "0.1.0"

type appConfig struct {
	addr string
	env  string
	db   dbConfig
}

type dbConfig struct {
	user         string
	password     string
	host         string
	port         int
	serviceName  string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

func main() {
	_ = godotenv.Load()

	cfg := appConfig{
		addr: env.GetString("APP_HOST", ":4000") + ":" + env.GetString("APP_PORT", "8080"),
		env:  env.GetString("APP_ENV", "development"),
		db: dbConfig{
			user:         env.GetString("ORACLE_USER", ""),
			password:     env.GetString("ORACLE_PASSWORD", ""),
			host:         env.GetString("ORACLE_HOST", "localhost"),
			port:         env.GetInt("ORACLE_PORT", 1521),
			serviceName:  env.GetString("ORACLE_SERVICE_NAME", "ORCLCDB.localdomain"),
			maxOpenConns: env.GetInt("ORACLE_MAX_OPEN_CONNS", 25),
			maxIdleConns: env.GetInt("ORACLE_MAX_IDLE_CONNS", 10),
			maxIdleTime:  env.GetString("ORACLE_MAX_IDLE_TIME", "15m"),
		},
	}

	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	dsn := "oracle://" + cfg.db.user + ":" + cfg.db.password + "@" + cfg.db.host + ":" + strconv.Itoa(cfg.db.port) + "/" + cfg.db.serviceName

	conn, err := db.New(
		dsn,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)

	if err != nil {
		logger.Fatalf("Error connecting to the database: %v", err)
	}
	defer conn.Close()

	logger.Infow("Connected to the database successfully")

	expvar.NewString("version").Set(version)
	expvar.Publish("database", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))

	logger.Infof("Starting server on %s in %s mode", cfg.addr, cfg.env)

	select {}
}
