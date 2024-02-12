package gorm_fixtures

import (
	"context"
	"fmt"

	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

type Config struct {
	ShowProgressBar     bool
	ResetAutoIncrements bool
	TruncateAllTables   bool
}

// Fixture представляет собой фикстуру для загрузки в базу данных.
type Fixture interface {
	Load(c *LoadCtx, db *gorm.DB) error
	Name() string
}

// DependentFixture представляет собой фикстуру, которая имеет зависимости от других фикстур.
type DependentFixture interface {
	Fixture
	GetRequiredRelations() []Fixture
}

type FixtureLoader struct {
	db       *gorm.DB
	fixtures []Fixture
	cleaner  *Cleaner
}

func NewFixtureLoader(db *gorm.DB, fixtures ...Fixture) *FixtureLoader {
	return &FixtureLoader{
		db:       db,
		fixtures: fixtures,
		cleaner:  NewCleaner(db),
	}
}

func (fl *FixtureLoader) Load(ctx context.Context, cfg Config) error {
	fixtures := fl.getFixtures()

	var bar *progressbar.ProgressBar
	if cfg.ShowProgressBar {
		bar = createProgressBar(len(fixtures))
	}

	if cfg.TruncateAllTables {
		err := fl.cleaner.TruncateAllTables()
		if err != nil {
			return fmt.Errorf("truncate all tables: %w", err)
		}
	}

	if cfg.ResetAutoIncrements {
		err := fl.cleaner.ResetAutoIncrementsCounters()
		if err != nil {
			return fmt.Errorf("reset auto increments: %w", err)
		}
	}

	loadCtx := NewLoadCtx(ctx)

	for _, fixture := range fixtures {
		if cfg.ShowProgressBar {
			bar.Describe(fmt.Sprintf("[INFO] Loading fixture: %s", fixture.Name()))
		}

		if err := fixture.Load(loadCtx, fl.db); err != nil {
			return fmt.Errorf("loading fixture %s: %w", fixture.Name(), err)
		}

		if cfg.ShowProgressBar {
			_ = bar.Add(1)
		}
	}
	return nil
}

func (fl *FixtureLoader) getFixtures() []Fixture {
	// Проверяем, реализует ли фикстура DependentFixture
	hasDependencies := false
	for _, fixture := range fl.fixtures {
		_, ok := fixture.(DependentFixture)
		if ok {
			hasDependencies = true
			break
		}
	}

	// Если есть фикстуры с зависимостями, используем функционал порядка загрузки
	if hasDependencies {
		// Создаем карту зависимостей между фикстурами
		dependencyMap := make(map[Fixture]bool)
		for _, fixture := range fl.fixtures {
			dependentFixture, ok := fixture.(DependentFixture)
			if ok {
				dependencyMap[fixture] = true
				for _, requiredFixture := range dependentFixture.GetRequiredRelations() {
					dependencyMap[requiredFixture] = true
				}
			}
		}

		// Определяем порядок загрузки фикстур
		var sortedFixtures []Fixture
		visited := make(map[Fixture]bool)
		var visit func(Fixture)
		visit = func(fixture Fixture) {
			if !visited[fixture] {
				visited[fixture] = true
				dependentFixture, ok := fixture.(DependentFixture)
				if ok {
					for _, requiredFixture := range dependentFixture.GetRequiredRelations() {
						visit(requiredFixture)
					}
				}
				sortedFixtures = append(sortedFixtures, fixture)
			}
		}
		for fixture := range dependencyMap {
			visit(fixture)
		}

		return sortedFixtures
	}

	return fl.fixtures
}
