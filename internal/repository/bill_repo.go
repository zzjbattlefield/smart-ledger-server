package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"smart-ledger-server/internal/model"
)

// BillRepository 账单数据访问层
type BillRepository struct {
	db *gorm.DB
}

// NewBillRepository 创建账单仓库
func NewBillRepository(db *gorm.DB) *BillRepository {
	return &BillRepository{db: db}
}

// Create 创建账单
func (r *BillRepository) Create(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Create(bill).Error
}

// GetByID 根据ID获取账单
func (r *BillRepository) GetByID(ctx context.Context, id uint64) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.WithContext(ctx).
		Preload("Category").
		First(&bill, id).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

// GetByUUID 根据UUID获取账单
func (r *BillRepository) GetByUUID(ctx context.Context, uuid string) (*model.Bill, error) {
	var bill model.Bill
	err := r.db.WithContext(ctx).
		Preload("Category").
		Where("uuid = ?", uuid).First(&bill).Error
	if err != nil {
		return nil, err
	}
	return &bill, nil
}

// BillQuery 账单查询条件
type BillQuery struct {
	UserID     uint64
	StartDate  *time.Time
	EndDate    *time.Time
	CategoryID *uint64
	BillType   *int
	Keyword    string
	Page       int
	PageSize   int
}

// List 查询账单列表
func (r *BillRepository) List(ctx context.Context, query *BillQuery) ([]model.Bill, int64, error) {
	var bills []model.Bill
	var total int64

	db := r.db.WithContext(ctx).Model(&model.Bill{}).Where("user_id = ?", query.UserID)

	// 时间范围
	if query.StartDate != nil {
		db = db.Where("pay_time >= ?", query.StartDate)
	}
	if query.EndDate != nil {
		db = db.Where("pay_time <= ?", query.EndDate)
	}

	// 分类
	if query.CategoryID != nil {
		db = db.Where("category_id = ?", *query.CategoryID)
	}

	// 账单类型
	if query.BillType != nil {
		db = db.Where("bill_type = ?", *query.BillType)
	}

	// 关键词搜索
	if query.Keyword != "" {
		keyword := "%" + query.Keyword + "%"
		db = db.Where("merchant LIKE ? OR remark LIKE ?", keyword, keyword)
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (query.Page - 1) * query.PageSize
	err := db.
		Preload("Category", "user_id = ?", query.UserID).
		Order("pay_time DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&bills).Error

	return bills, total, err
}

// Update 更新账单
func (r *BillRepository) Update(ctx context.Context, bill *model.Bill) error {
	return r.db.WithContext(ctx).Save(bill).Error
}

// Delete 删除账单(软删除)
func (r *BillRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&model.Bill{}, id).Error
}

// StatsSummary 统计结果
type StatsSummary struct {
	TotalExpense decimal.Decimal
	TotalIncome  decimal.Decimal
	BillCount    int64
}

// GetStatsSummary 获取统计摘要
func (r *BillRepository) GetStatsSummary(ctx context.Context, userID uint64, startDate, endDate time.Time) (*StatsSummary, error) {
	var result StatsSummary

	// 统计支出
	var expense decimal.Decimal
	err := r.db.WithContext(ctx).Model(&model.Bill{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND bill_type = ? AND pay_time >= ? AND pay_time <= ?",
			userID, model.BillTypeExpense, startDate, endDate).
		Scan(&expense).Error
	if err != nil {
		return nil, err
	}
	result.TotalExpense = expense

	// 统计收入
	var income decimal.Decimal
	err = r.db.WithContext(ctx).Model(&model.Bill{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND bill_type = ? AND pay_time >= ? AND pay_time <= ?",
			userID, model.BillTypeIncome, startDate, endDate).
		Scan(&income).Error
	if err != nil {
		return nil, err
	}
	result.TotalIncome = income

	// 统计数量
	err = r.db.WithContext(ctx).Model(&model.Bill{}).
		Where("user_id = ? AND pay_time >= ? AND pay_time <= ?", userID, startDate, endDate).
		Count(&result.BillCount).Error
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// CategoryStats 分类统计结果
type CategoryStats struct {
	CategoryID   uint64
	CategoryName string
	Amount       decimal.Decimal
}

// GetCategoryStats 获取分类统计
func (r *BillRepository) GetCategoryStats(ctx context.Context, userID uint64, billType model.BillType, startDate, endDate time.Time) ([]CategoryStats, error) {
	var stats []CategoryStats
	err := r.db.WithContext(ctx).Model(&model.Bill{}).
		Select("category_id, categories.name as category_name, SUM(bills.amount) as amount").
		Joins("LEFT JOIN categories ON bills.category_id = categories.id AND categories.user_id = bills.user_id").
		Where("bills.user_id = ? AND bills.bill_type = ? AND bills.pay_time >= ? AND bills.pay_time <= ?",
			userID, billType, startDate, endDate).
		Group("category_id").
		Group("category_name").
		Order("amount DESC").
		Scan(&stats).Error
	return stats, err
}

// DailyStats 每日统计结果
type DailyStats struct {
	Date    time.Time
	Expense decimal.Decimal
	Income  decimal.Decimal
}

// GetDailyStats 获取每日统计
func (r *BillRepository) GetDailyStats(ctx context.Context, userID uint64, startDate, endDate time.Time) ([]DailyStats, error) {
	var stats []DailyStats
	err := r.db.WithContext(ctx).Model(&model.Bill{}).
		Select(`
			DATE(pay_time) as date,
			SUM(CASE WHEN bill_type = 1 THEN amount ELSE 0 END) as expense,
			SUM(CASE WHEN bill_type = 2 THEN amount ELSE 0 END) as income
		`).
		Where("user_id = ? AND pay_time >= ? AND pay_time <= ?", userID, startDate, endDate).
		Group("DATE(pay_time)").
		Order("date ASC").
		Scan(&stats).Error
	return stats, err
}
