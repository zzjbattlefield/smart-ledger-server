package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upUserCategory, downUserCategory)
}

func upUserCategory(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	if _, err := tx.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS category_templates (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			parent_id BIGINT UNSIGNED DEFAULT 0,
			icon VARCHAR(100),
			sort_order INT DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			INDEX idx_parent_id (parent_id),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO category_templates (id, name, parent_id, icon, sort_order, created_at, updated_at, deleted_at)
		SELECT id, name, parent_id, icon, sort_order, created_at, updated_at, deleted_at FROM categories
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM categories`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories
			ADD COLUMN user_id BIGINT UNSIGNED NOT NULL AFTER id,
			DROP COLUMN is_system,
			ADD INDEX idx_user_id (user_id)
	`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories ADD UNIQUE INDEX uk_user_parent_name (user_id, parent_id, name)
	`); err != nil {
		return err
	}

	rows, err := tx.QueryContext(ctx, `SELECT id FROM users WHERE deleted_at IS NULL`)
	if err != nil {
		return err
	}
	defer rows.Close()
	userIDs := make([]uint64, 0)
	for rows.Next() {
		var userID uint64
		if err := rows.Scan(&userID); err != nil {
			return err
		}
		userIDs = append(userIDs, userID)
	}
	for _, id := range userIDs {
		if err = initCategoriesForUser(ctx, tx, id); err != nil {
			return err
		}
	}
	return nil
}

func initCategoriesForUser(ctx context.Context, tx *sql.Tx, userID uint64) error {
	rows, err := tx.QueryContext(ctx, `SELECT id, name, parent_id, icon, sort_order FROM category_templates WHERE deleted_at IS NULL ORDER BY id`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type template struct {
		ID        uint64
		Name      string
		ParentID  uint64
		Icon      string
		SortOrder int
	}
	var templates []template
	for rows.Next() {
		var t template
		if err = rows.Scan(&t.ID, &t.Name, &t.ParentID, &t.Icon, &t.SortOrder); err != nil {
			return err
		}
		templates = append(templates, t)
	}
	idMap := make(map[uint64]uint64)
	// 先插入顶级分类（parent_id = 0）
	for _, t := range templates {
		if t.ParentID == 0 {
			result, err := tx.ExecContext(ctx,
				"INSERT INTO categories (user_id, name, parent_id, icon, sort_order) VALUES (?, ?, 0, ?, ?)",
				userID, t.Name, t.Icon, t.SortOrder,
			)
			if err != nil {
				return err
			}
			newID, _ := result.LastInsertId()
			idMap[t.ID] = uint64(newID)
		}
	}

	// 再插入子分类（映射 parent_id）
	for _, t := range templates {
		if t.ParentID != 0 {
			newParentID := idMap[t.ParentID]
			result, err := tx.ExecContext(ctx,
				"INSERT INTO categories (user_id, name, parent_id, icon, sort_order) VALUES (?, ?, ?, ?, ?)",
				userID, t.Name, newParentID, t.Icon, t.SortOrder,
			)
			if err != nil {
				return err
			}
			newID, _ := result.LastInsertId()
			idMap[t.ID] = uint64(newID)
		}
	}

	// 兼容已有账单数据：把旧分类ID（模板ID）映射为新分类ID
	for oldID, newID := range idMap {
		if _, err := tx.ExecContext(ctx,
			"UPDATE bills SET category_id = ? WHERE user_id = ? AND category_id = ?",
			newID, userID, oldID,
		); err != nil {
			return err
		}
	}

	return nil
}

func downUserCategory(ctx context.Context, tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
