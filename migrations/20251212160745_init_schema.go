package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upInitSchema, downInitSchema)
}

func upInitSchema(ctx context.Context, tx *sql.Tx) error {
	// 创建用户表
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			phone VARCHAR(20) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			nickname VARCHAR(50),
			avatar_url VARCHAR(255),
			last_login_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			INDEX idx_phone (phone),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}

	// 创建分类表
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS categories (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(50) NOT NULL,
			parent_id BIGINT UNSIGNED DEFAULT 0,
			icon VARCHAR(100),
			sort_order INT DEFAULT 0,
			is_system TINYINT(1) DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			INDEX idx_parent_id (parent_id),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}

	// 创建账单表
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS bills (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			uuid VARCHAR(36) NOT NULL UNIQUE,
			user_id BIGINT UNSIGNED NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			bill_type TINYINT DEFAULT 1 COMMENT '1:支出 2:收入',
			platform VARCHAR(50),
			merchant VARCHAR(255),
			category_id BIGINT UNSIGNED,
			pay_time DATETIME NOT NULL,
			pay_method VARCHAR(50),
			order_no VARCHAR(100),
			remark VARCHAR(500),
			image_path VARCHAR(255),
			ai_raw_response TEXT,
			confidence DECIMAL(3,2),
			is_confirmed TINYINT(1) DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			deleted_at DATETIME,
			INDEX idx_user_id (user_id),
			INDEX idx_category_id (category_id),
			INDEX idx_pay_time (pay_time),
			INDEX idx_deleted_at (deleted_at)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
	`); err != nil {
		return err
	}

	// 插入默认分类
	categories := []struct {
		Name      string
		ParentID  int
		Icon      string
		SortOrder int
	}{
		// 餐饮
		{"餐饮", 0, "restaurant", 1},
		{"正餐", 1, "lunch", 1},
		{"小吃零食", 1, "fastfood", 2},
		{"咖啡饮品", 1, "coffee", 3},
		{"水果生鲜", 1, "fruit", 4},
		{"外卖配送费", 1, "delivery", 5},
		// 交通
		{"交通", 0, "transport", 2},
		{"公共交通", 7, "bus", 1},
		{"打车", 7, "taxi", 2},
		{"共享单车", 7, "bike", 3},
		{"加油停车", 7, "gas", 4},
		// 购物
		{"购物", 0, "shopping", 3},
		{"日用百货", 12, "daily", 1},
		{"服饰鞋包", 12, "clothing", 2},
		{"数码电器", 12, "digital", 3},
		{"美妆护肤", 12, "cosmetics", 4},
		// 娱乐
		{"娱乐", 0, "entertainment", 4},
		{"电影演出", 17, "movie", 1},
		{"游戏充值", 17, "game", 2},
		{"会员订阅", 17, "subscription", 3},
		{"运动健身", 17, "sports", 4},
		// 生活服务
		{"生活服务", 0, "life", 5},
		{"话费充值", 22, "phone", 1},
		{"水电燃气", 22, "utility", 2},
		{"医疗健康", 22, "medical", 3},
		{"快递物流", 22, "express", 4},
		{"其他服务", 22, "other", 5},
		// 金融
		{"金融", 0, "finance", 6},
		{"转账", 28, "transfer", 1},
		{"还款", 28, "repayment", 2},
		{"理财", 28, "investment", 3},
		{"保险", 28, "insurance", 4},
	}

	for _, c := range categories {
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO categories (name, parent_id, icon, sort_order, is_system) VALUES (?, ?, ?, ?, 1)",
			c.Name, c.ParentID, c.Icon, c.SortOrder,
		); err != nil {
			return err
		}
	}

	return nil
}

func downInitSchema(ctx context.Context, tx *sql.Tx) error {
	tables := []string{"bills", "categories", "users"}
	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, "DROP TABLE IF EXISTS "+table); err != nil {
			return err
		}
	}
	return nil
}
