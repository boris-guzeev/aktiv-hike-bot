package config

import (
	"database/sql"
	"log"
	"os"
	"strconv"
)

type Config struct {
	AdminBotToken string
	AdminChatID   int64
	DatabaseURL   string
}

func MustLoad() Config {
	adminChat := mustParseInt64(getenv("ADMIN_CHAT_ID"))
	return Config{
		AdminBotToken: getenv("ADMIN_BOT_TOKEN"),
		AdminChatID:   adminChat,
		DatabaseURL:   getenv("DB_DSN"),
	}
}

func getenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}

func mustParseInt64(s string) int64 {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatalf("bad int64: %v", err)
	}
	return int64(i)
}

// TODO: помоему плохая идея в конфигах делать инициализацию объекта БД
func MustOpenDB(url string) *sql.DB {
	db, err := sql.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	return db
}
