package main

import (
	"github.com/joho/godotenv"
	"github.com/quangtran666/simple-social-golang/internal/db"
	"github.com/quangtran666/simple-social-golang/internal/env"
	"github.com/quangtran666/simple-social-golang/internal/mailer"
	"github.com/quangtran666/simple-social-golang/internal/store"
	"go.uber.org/zap"
	"log"
	"time"
)

const version = "0.0.1"

//	@title			GopherSocial API
//	@description	This is a simple social media API
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@BasePath	/v1

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				JWT Authorization header using Bearer scheme
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file, error: %v", err)
	}

	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgresql://postgres:postgres@localhost:5432/simple_social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
		mail: mailConfig{
			exp: time.Hour * 24 * 3,
			mailtrap: mailtrapConfig{
				fromEmail: env.GetString("MAIL_FROM_EMAIL", "test@gmail.com"),
				username:  env.GetString("MAIL_USERNAME", "test"),
				password:  env.GetString("MAIL_PASSWORD", "test"),
			},
		},
	}

	// Logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	// Database
	db, err := db.New(cfg.db.addr, cfg.db.maxOpenConns, cfg.db.maxIdleConns, cfg.db.maxIdleTime)
	if err != nil {
		log.Fatalf("Error establishing database connection, error: %v", err)
	}
	defer db.Close()
	logger.Info("database connection pool has been established")
	store := store.NewStorage(db)

	mailer := mailer.NewMailtrapMailer(cfg.mail.mailtrap.fromEmail, cfg.mail.mailtrap.username, cfg.mail.mailtrap.password)

	app := &application{
		config: cfg,
		store:  store,
		logger: logger,
		mailer: mailer,
	}

	mux := app.mount()
	logger.Fatal(app.run(mux))
}
