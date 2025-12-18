package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upAddCategoryType, downAddCategoryType)
}

func upAddCategoryType(ctx context.Context, tx *sql.Tx) error {
	// 1. 为 category_templates 添加 type 字段
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE category_templates
		ADD COLUMN type TINYINT NOT NULL DEFAULT 1 COMMENT '1:支出 2:收入' AFTER name
	`); err != nil {
		return err
	}

	// 为 category_templates 添加索引
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE category_templates ADD INDEX idx_type (type)
	`); err != nil {
		return err
	}

	// 2. 为 categories 添加 type 字段
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories
		ADD COLUMN type TINYINT NOT NULL DEFAULT 1 COMMENT '1:支出 2:收入' AFTER name
	`); err != nil {
		return err
	}

	// 为 categories 添加索引
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories ADD INDEX idx_type (type)
	`); err != nil {
		return err
	}

	// 3. 删除旧唯一索引
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories DROP INDEX uk_user_parent_name
	`); err != nil {
		return err
	}

	// 4. 创建新唯一索引（包含 type）
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories ADD UNIQUE INDEX uk_user_parent_name_type (user_id, parent_id, name, type)
	`); err != nil {
		return err
	}

	// 5. 插入收入类型模板
	// 先插入一级分类
	incomeTopCategories := []struct {
		Name      string
		Icon      string
		SortOrder int
	}{
		{"薪资", "salary", 1},
		{"收红包", "red_packet", 2},
		{"奖金", "bonus", 3},
		{"其他", "other_income", 4},
	}

	var otherCategoryID int64
	for _, cat := range incomeTopCategories {
		result, err := tx.ExecContext(ctx,
			"INSERT INTO category_templates (name, type, parent_id, icon, sort_order) VALUES (?, 2, 0, ?, ?)",
			cat.Name, cat.Icon, cat.SortOrder,
		)
		if err != nil {
			return err
		}
		if cat.Name == "其他" {
			otherCategoryID, _ = result.LastInsertId()
		}
	}

	// 插入二级分类（挂在"其他"下面）
	incomeSubCategories := []struct {
		Name      string
		Icon      string
		SortOrder int
	}{
		{"公积金", "housing_fund", 1},
		{"意外来财", "windfall", 2},
	}

	for _, cat := range incomeSubCategories {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO category_templates (name, type, parent_id, icon, sort_order) VALUES (?, 2, ?, ?, ?)",
			cat.Name, otherCategoryID, cat.Icon, cat.SortOrder,
		); err != nil {
			return err
		}
	}

	// 6. 为现有用户初始化收入分类
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

	for _, userID := range userIDs {
		if err := initIncomeCategoriesForUser(ctx, tx, userID); err != nil {
			return err
		}
	}

	return nil
}

// initIncomeCategoriesForUser 为用户初始化收入分类（只处理 type=2 的模板）
func initIncomeCategoriesForUser(ctx context.Context, tx *sql.Tx, userID uint64) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, name, parent_id, icon, sort_order FROM category_templates WHERE type = 2 AND deleted_at IS NULL ORDER BY id`)
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
		if err := rows.Scan(&t.ID, &t.Name, &t.ParentID, &t.Icon, &t.SortOrder); err != nil {
			return err
		}
		templates = append(templates, t)
	}

	idMap := make(map[uint64]uint64)

	// 先插入顶级分类（parent_id = 0）
	for _, t := range templates {
		if t.ParentID == 0 {
			result, err := tx.ExecContext(ctx,
				"INSERT INTO categories (user_id, name, type, parent_id, icon, sort_order) VALUES (?, ?, 2, 0, ?, ?)",
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
				"INSERT INTO categories (user_id, name, type, parent_id, icon, sort_order) VALUES (?, ?, 2, ?, ?, ?)",
				userID, t.Name, newParentID, t.Icon, t.SortOrder,
			)
			if err != nil {
				return err
			}
			newID, _ := result.LastInsertId()
			idMap[t.ID] = uint64(newID)
		}
	}

	return nil
}

func downAddCategoryType(ctx context.Context, tx *sql.Tx) error {
	// 回滚：删除收入分类数据、删除字段、恢复索引
	// 1. 删除用户的收入分类
	if _, err := tx.ExecContext(ctx, `DELETE FROM categories WHERE type = 2`); err != nil {
		return err
	}

	// 2. 删除收入类型模板
	if _, err := tx.ExecContext(ctx, `DELETE FROM category_templates WHERE type = 2`); err != nil {
		return err
	}

	// 3. 删除新唯一索引
	if _, err := tx.ExecContext(ctx, `ALTER TABLE categories DROP INDEX uk_user_parent_name_type`); err != nil {
		return err
	}

	// 4. 恢复旧唯一索引
	if _, err := tx.ExecContext(ctx, `
		ALTER TABLE categories ADD UNIQUE INDEX uk_user_parent_name (user_id, parent_id, name)
	`); err != nil {
		return err
	}

	// 5. 删除 categories 的 type 索引和字段
	if _, err := tx.ExecContext(ctx, `ALTER TABLE categories DROP INDEX idx_type`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE categories DROP COLUMN type`); err != nil {
		return err
	}

	// 6. 删除 category_templates 的 type 索引和字段
	if _, err := tx.ExecContext(ctx, `ALTER TABLE category_templates DROP INDEX idx_type`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE category_templates DROP COLUMN type`); err != nil {
		return err
	}

	return nil
}
