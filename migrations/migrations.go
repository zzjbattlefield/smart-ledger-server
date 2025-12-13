package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

// Run 执行数据库迁移
func Run(db *sql.DB) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	return goose.Up(db, ".")
}

// Rollback 回滚最近一次迁移
func Rollback(db *sql.DB) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	return goose.Down(db, ".")
}

// Status 显示迁移状态
func Status(db *sql.DB) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}

	return goose.Status(db, ".")
}
