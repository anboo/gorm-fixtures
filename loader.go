package gorm_fixtures

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
	"gorm.io/gorm"
)

// Fixture представляет собой фикстуру для загрузки в базу данных.
type Fixture interface {
	Load(db *gorm.DB) error
	Name() string
}

// DependentFixture представляет собой фикстуру, которая имеет зависимости от других фикстур.
type DependentFixture interface {
	Fixture
	GetRequiredRelations() []Fixture
}

type FixtureLoader struct {
	DB       *gorm.DB
	Fixtures []Fixture
}

func NewFixtureLoader(db *gorm.DB, fixtures ...Fixture) *FixtureLoader {
	return &FixtureLoader{
		DB:       db,
		Fixtures: fixtures,
	}
}

func (fl *FixtureLoader) Load() error {
	fixtures := fl.getFixtures()

	totalFixtures := len(fixtures)
	bar := progressbar.NewOptions(totalFixtures,
		progressbar.OptionSetWriter(ansiWriter{}),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetDescription("[INFO] Loading fixtures"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionEnableColorCodes(true),
	)
	for _, fixture := range fixtures {
		bar.Describe(fmt.Sprintf("[INFO] Loading fixture: %s", fixture.Name()))
		if err := fixture.Load(fl.DB); err != nil {
			return err
		}
		bar.Add(1)
	}
	return nil
}

func (fl *FixtureLoader) getFixtures() []Fixture {
	// Проверяем, реализует ли фикстура DependentFixture
	hasDependencies := false
	for _, fixture := range fl.Fixtures {
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
		for _, fixture := range fl.Fixtures {
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

	return fl.Fixtures
}

type ansiWriter struct{}

// Write записывает данные в терминал с использованием ANSI цветов.
func (w ansiWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(string(p))
}
