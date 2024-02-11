package gorm_fixtures

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Cleaner struct {
	db *gorm.DB
}

func (c *Cleaner) ResetAutoIncrementsCounters() error {
	tables, err := c.getTables()
	if err != nil {
		return err
	}

	for _, table := range tables {
		switch {
		case c.isMySQL():
			if err := c.db.Exec("ALTER TABLE " + table + " AUTO_INCREMENT = 1").Error; err != nil {
				return fmt.Errorf("restart auto_increment for table %s: %w", table, err)
			}
		case c.isPostgreSQL():
			if err := c.db.Exec("ALTER SEQUENCE " + table + "_id_seq RESTART WITH 1").Error; err != nil {
				return fmt.Errorf("restart seq for table %s: %w", table, err)
			}
		}
	}

	return nil
}

func (c *Cleaner) TruncateAllTables() error {
	// Получаем список всех таблиц в базе данных
	tables, err := c.getTables()
	if err != nil {
		return err
	}

	// Собираем мапу с зависимостями между таблицами
	dependencies := make(map[string][]string)
	for _, table := range tables {
		var refs []string
		if err := c.db.Raw("SELECT referenced_table_name FROM information_schema.key_column_usage WHERE table_name = ? AND table_schema = ?", table, c.db.Migrator().CurrentDatabase()).Pluck("referenced_table_name", &refs).Error; err != nil {
			return err
		}
		dependencies[table] = refs
	}

	// Определяем порядок удаления таблиц с учетом зависимостей
	var deletionOrder []string
	for len(dependencies) > 0 {
		var noDependencies []string
		for table, refs := range dependencies {
			if len(refs) == 0 {
				noDependencies = append(noDependencies, table)
				delete(dependencies, table)
			}
		}
		if len(noDependencies) == 0 {
			return fmt.Errorf("cyclic dependencies detected, unable to truncate all tables")
		}
		deletionOrder = append(deletionOrder, noDependencies...)
		for table := range dependencies {
			for _, dep := range noDependencies {
				for i, ref := range dependencies[table] {
					if ref == dep {
						dependencies[table] = append(dependencies[table][:i], dependencies[table][i+1:]...)
						break
					}
				}
			}
		}
	}

	// Очищаем таблицы в порядке удаления
	for _, table := range deletionOrder {
		if err := c.db.Exec("TRUNCATE TABLE " + table).Error; err != nil {
			return err
		}
	}

	return nil
}

func (c *Cleaner) getTables() ([]string, error) {
	var tables []string
	if err := c.db.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = ?", c.db.Migrator().CurrentDatabase()).Pluck("table_name", &tables).Error; err != nil {
		return nil, err
	}
	return tables, nil
}

// IsMySQL возвращает true, если используется MySQL, иначе - false.
func (c *Cleaner) isMySQL() bool {
	_, ok := c.db.Dialector.(*mysql.Dialector)
	return ok
}

// IsPostgreSQL возвращает true, если используется PostgreSQL, иначе - false.
func (c *Cleaner) isPostgreSQL() bool {
	_, ok := c.db.Dialector.(*postgres.Dialector)
	return ok
}
