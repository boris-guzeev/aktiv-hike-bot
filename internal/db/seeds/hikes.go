package seeds

import (
	"context"
	"fmt"
	"time"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type Seeder struct {
	q   *sqlc.Queries
	loc *time.Location
}

func New(q *sqlc.Queries, loc *time.Location) *Seeder {
	return &Seeder{q: q, loc: loc}
}

func (s *Seeder) Seed(ctx context.Context) error {
	hikes := []sqlc.CreateHikeParams{
		// 1) Однодневный
		{
			TitleRu: "Мтирала — водопады и туманная тропа",
			DescriptionRu: `Маршрут по национальному парку Мтирала с мягким набором высоты, 
каскадными водопадами и туманным субтропическим лесом. Подойдёт для тех, кто хочет 
познакомиться с хайкингом без сильных нагрузок, но с красивыми видами.`,
			TitleEn:       pgtype.Text{}, // пусто
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2025, 12, 7, 9, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2025, 12, 7, 19, 0, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
		},

		// 2) Однодневный
		{
			TitleRu: "Махунцети — мост царицы Тамары и водопад",
			DescriptionRu: `Короткий, но насыщенный маршрут: древний арочный мост, 
виды на горную Аджарию и прогулка к водопаду. Отличный вариант для лёгкой вылазки 
или знакомства с форматом хайков.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2025, 12, 10, 10, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2025, 12, 10, 16, 30, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
		},

		// 3) Двухдневный
		{
			TitleRu: "Чирокхи → Муха — хребет над облаками (2 дня)",
			DescriptionRu: `Двухдневный маршрут по хребту с ночёвкой в палатках. 
Открытые панорамы, звёздное небо и мягкие переходы. Хороший вариант для первого 
двухдневного похода.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2025, 12, 13, 8, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2025, 12, 14, 18, 0, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
		},

		// 4) Трёхдневный
		{
			TitleRu: "Аджарский хребет — мини-экспедиция (3 дня)",
			DescriptionRu: `Усиленный маршрут с набором высоты и длинными переходами. 
За три дня мы пройдём по хребту, увидим смену ландшафтов и почувствуем настоящий 
походный ритм с ночёвками в палатках.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2025, 12, 19, 7, 30, 0, 0, s.loc),
			EndsAt:        time.Date(2025, 12, 21, 20, 0, 0, 0, s.loc),
			IsPublished:   false,
			CreatedAt:     time.Now(),
		},

		// 5) Четырёхдневный
		{
			TitleRu: "Тамаршени и заброшенные пастбища (4 дня)",
			DescriptionRu: `Маршрут для любителей уединения и тихих ландшафтов. 
Мы проходим через заброшенные сёла и пастбища, ночуем в палатках и много времени 
уделяем фотографиям и созерцанию.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2025, 12, 26, 9, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2025, 12, 29, 18, 0, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
		},

		// 6) Однодневный
		{
			TitleRu: "Сарпи → Хирс — тропа вдоль моря",
			DescriptionRu: `Красивейший маршрут вдоль Черного моря: скальные участки, 
прибрежные тропы и широкие виды на Турцию. Идеален для межсезонья.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2026, 1, 4, 9, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2026, 1, 4, 18, 30, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
		},

		// 7) Недельный поход (7 дней)
		{
			TitleRu: "Большое кольцо Аджарского хребта (7 дней)",
			DescriptionRu: `Полноценный недельный поход с переходами, перевалами, 
палаточным лагерем и полным отключением от городской суеты. Формат — трек 
для тех, кто хочет прочувствовать ритм гор и восстановить голову.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2026, 1, 12, 9, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2026, 1, 19, 18, 0, 0, 0, s.loc),
			IsPublished:   false,
			CreatedAt:     time.Now(),
		},

		// 8) Длинный маршрут 8 дней
		{
			TitleRu: "Черноморский трейл: Батуми → Кобулети → горы (8 дней)",
			DescriptionRu: `Комбинированный маршрут: побережье, тропы вдоль моря, 
маленькие сёла и переход в горную часть с ночёвками в палатках. Подойдёт тем, 
кто любит путешествия, а не просто точки на карте.`,
			TitleEn:       pgtype.Text{},
			DescriptionEn: pgtype.Text{},
			PhotoFileID:   pgtype.Text{},
			StartsAt:      time.Date(2026, 2, 2, 9, 0, 0, 0, s.loc),
			EndsAt:        time.Date(2026, 2, 10, 18, 0, 0, 0, s.loc),
			IsPublished:   true,
			CreatedAt:     time.Now(),
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
