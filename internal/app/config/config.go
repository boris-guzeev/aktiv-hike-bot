package config

import (
	"database/sql"
	"log"
	"os"
	"strconv"
)

type Common struct {
	AdminChatID int64
	DatabaseURL string
	StorageRoot string
	Timezone    string
}

type AdminBot struct {
	Common
	AdminBotToken string
}

type ClientBot struct {
	Common
	ClientBotToken string
	AdminBotName   string
}

func MustLoadCommon() Common {
	adminChat := mustParseInt64(getenv("ADMIN_CHAT_ID"))
	return Common{
		AdminChatID: adminChat,
		DatabaseURL: getenv("DB_DSN"),
		StorageRoot: os.Getenv("STORAGE_ROOT"),
		Timezone:    os.Getenv("TZ"),
	}
}

func MustLoadAdminBot() AdminBot {
	common := MustLoadCommon()
	return AdminBot{
		Common:        common,
		AdminBotToken: getenv("ADMIN_BOT_TOKEN"),
	}
}

func MustLoadClientBot() ClientBot {
	common := MustLoadCommon()
	return ClientBot{
		Common:         common,
		ClientBotToken: getenv("CLIENT_BOT_TOKEN"),
		AdminBotName:   getenv("ADMIN_BOT_NAME"),
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
