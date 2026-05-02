package seeds

import (
	"context"
	"fmt"
	"math/big"
	"time"

	sqlc "github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc/admin"
	"github.com/jackc/pgx/v5/pgtype"
)

type Seeder struct {
	q   *sqlc.Queries
	loc *time.Location
}

func New(q *sqlc.Queries, loc *time.Location) *Seeder {
	return &Seeder{q: q, loc: loc}
}

func numeric(v int64, exp int32) pgtype.Numeric {
	return pgtype.Numeric{Int: big.NewInt(v), Exp: exp, Valid: true}
}

func (s *Seeder) Seed(ctx context.Context) error {
	now := time.Now().In(s.loc)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)

	longDescriptionRu := `Это подробное описание хайка.

Мы заранее встречаемся в удобной точке, знакомимся с группой, коротко обсуждаем маршрут, темп движения и правила безопасности. По пути делаем остановки для отдыха, фотографий и небольших перекусов. Маршрут рассчитан так, чтобы сохранить баланс между активной нагрузкой и удовольствием от природы.

Во время хайка нас ждут красивые виды, лесные тропы, горный воздух и спокойный ритм движения. Мы не спешим, держимся вместе и ориентируемся на общий уровень группы. Если маршрут проходит несколько дней, ночёвка организуется в палатках, а вечером остаётся время на отдых, разговоры и восстановление после перехода.

С собой желательно взять удобную обувь, воду, перекус, дождевик, тёплый слой одежды и хорошее настроение. Перед стартом мы дополнительно напомним список вещей и уточним детали по погоде, трансферу и времени встречи.

Этот хайк подойдёт тем, кто хочет выбраться из города, перезагрузиться, увидеть новые места и провести день или несколько дней в компании людей, которым тоже нравится природа, движение и горы.`

	hikes := []sqlc.CreateHikeParams{
		{
			TitleRu:        "Мтирала — водопады и туманная тропа",
			PreviewRu:      `Лёгкий маршрут по Мтирале: водопады, туманный лес и мягкий набор высоты.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       80,
			ElevationGainM: pgtype.Int4{Int32: 450, Valid: true},
			DistanceKm:     numeric(85, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 3)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 3)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
		{
			TitleRu:        "Махунцети — мост царицы Тамары и водопад",
			PreviewRu:      `Короткая вылазка к мосту и водопаду.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       70,
			ElevationGainM: pgtype.Int4{Int32: 250, Valid: true},
			DistanceKm:     numeric(60, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 10)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 10)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
		{
			TitleRu:        "Чирокхи → Муха (2 дня)",
			PreviewRu:      `Двухдневный поход по хребту с палатками.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       180,
			ElevationGainM: pgtype.Int4{Int32: 900, Valid: true},
			DistanceKm:     numeric(180, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 14)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 15)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
		{
			TitleRu:        "Аджарский хребет (3 дня)",
			PreviewRu:      `Интенсивный маршрут с набором высоты.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       260,
			ElevationGainM: pgtype.Int4{Int32: 1450, Valid: true},
			DistanceKm:     numeric(320, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 15)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 17)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: false,
		},
		{
			TitleRu:        "Тамаршени (4 дня)",
			PreviewRu:      `Уединённый маршрут по пастбищам.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       320,
			ElevationGainM: pgtype.Int4{Int32: 1800, Valid: true},
			DistanceKm:     numeric(420, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 19)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 22)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
		{
			TitleRu:        "Сарпи → Хирс",
			PreviewRu:      `Маршрут вдоль моря.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       90,
			ElevationGainM: pgtype.Int4{Int32: 500, Valid: true},
			DistanceKm:     numeric(110, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 23)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 23)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
		{
			TitleRu:        "Большое кольцо (7 дней)",
			PreviewRu:      `Недельный поход по горам.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       600,
			ElevationGainM: pgtype.Int4{Int32: 3600, Valid: true},
			DistanceKm:     numeric(760, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 25)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 31)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: false,
		},
		{
			TitleRu:        "Черноморский трейл (8 дней)",
			PreviewRu:      `Путешествие от моря в горы.`,
			DescriptionRu:  longDescriptionRu,
			TitleEn:        pgtype.Text{},
			DescriptionEn:  pgtype.Text{},
			ImagePath:      pgtype.Text{String: "hikes/3.jpg", Valid: true},
			PriceGel:       720,
			ElevationGainM: pgtype.Int4{Int32: 4200, Valid: true},
			DistanceKm:     numeric(880, -1),
			StartsAt: func() time.Time {
				d := today.AddDate(0, 0, 32)
				return time.Date(d.Year(), d.Month(), d.Day(), 8, 0, 0, 0, s.loc)
			}(),
			EndsAt: func() time.Time {
				d := today.AddDate(0, 0, 39)
				return time.Date(d.Year(), d.Month(), d.Day(), 22, 0, 0, 0, s.loc)
			}(),
			IsPublished: true,
		},
	}

	for _, h := range hikes {
		_, err := s.q.CreateHike(ctx, h)
		if err != nil {
			return fmt.Errorf("failed to seed hike %s: %w", h.TitleRu, err)
		}
	}

	return nil
}
