package service

import (
	"context"
	"time"

	"github.com/shopspring/decimal"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/pkg/errcode"
)

// StatsService 统计服务
type StatsService struct {
	billRepo BillRepo
}

// NewStatsService 创建统计服务
func NewStatsService(billRepo BillRepo) *StatsService {
	return &StatsService{
		billRepo: billRepo,
	}
}

// GetSummary 获取统计摘要
func (s *StatsService) GetSummary(ctx context.Context, userID uint64, req *dto.StatsSummaryRequest) (*dto.StatsSummaryResponse, error) {
	startDate, endDate, err := s.parsePeriod(req.Period, req.Date)
	if err != nil {
		return nil, errcode.ErrParams.WithMessage(err.Error())
	}

	// 获取基础统计
	summary, err := s.billRepo.GetStatsSummary(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 计算日均消费
	days := endDate.Sub(startDate).Hours() / 24
	if days < 1 {
		days = 1
	}
	dailyAverage := summary.TotalExpense.Div(decimal.NewFromFloat(days))

	// 获取分类统计
	categoryStats, err := s.billRepo.GetCategoryStats(ctx, userID, model.BillTypeExpense, startDate, endDate)
	if err != nil {
		return nil, errcode.ErrServer
	}

	topCategories := make([]dto.CategoryStatsItem, len(categoryStats))
	for i, stat := range categoryStats {
		percent := 0.0
		if !summary.TotalExpense.IsZero() {
			percent, _ = stat.Amount.Div(summary.TotalExpense).Mul(decimal.NewFromInt(100)).Float64()
		}
		topCategories[i] = dto.CategoryStatsItem{
			ID:      stat.CategoryID,
			Name:    stat.CategoryName,
			Amount:  stat.Amount.Round(2),
			Percent: percent,
		}
	}

	// 获取趋势数据
	dailyStats, err := s.billRepo.GetDailyStats(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, errcode.ErrServer
	}

	trend := make([]dto.TrendItem, len(dailyStats))
	for i, stat := range dailyStats {
		trend[i] = dto.TrendItem{
			Date:    dto.DateOnly(stat.Date),
			Expense: stat.Expense,
			Income:  stat.Income,
		}
	}

	return &dto.StatsSummaryResponse{
		Period:        req.Date,
		TotalExpense:  summary.TotalExpense.Round(2),
		TotalIncome:   summary.TotalIncome.Round(2),
		BillCount:     summary.BillCount,
		DailyAverage:  dailyAverage.Round(2),
		TopCategories: topCategories,
		Trend:         trend,
	}, nil
}

// GetCategoryStats 获取分类统计
func (s *StatsService) GetCategoryStats(ctx context.Context, userID uint64, req *dto.StatsCategoryRequest) (*dto.CategoryStatsResponse, error) {
	startDate, endDate, err := s.parsePeriod(req.Period, req.Date)
	if err != nil {
		return nil, errcode.ErrParams.WithMessage(err.Error())
	}

	// 获取分类统计
	categoryStats, err := s.billRepo.GetCategoryStats(ctx, userID, model.BillTypeExpense, startDate, endDate)
	if err != nil {
		return nil, errcode.ErrServer
	}

	// 计算总金额
	var total decimal.Decimal
	for _, stat := range categoryStats {
		total = total.Add(stat.Amount)
	}

	categories := make([]dto.CategoryStatsItem, len(categoryStats))
	for i, stat := range categoryStats {
		percent := 0.0
		if !total.IsZero() {
			percent, _ = stat.Amount.Div(total).Mul(decimal.NewFromInt(100)).Round(2).Float64()
		}
		categories[i] = dto.CategoryStatsItem{
			ID:      stat.CategoryID,
			Name:    stat.CategoryName,
			Amount:  stat.Amount,
			Percent: percent,
		}
	}

	return &dto.CategoryStatsResponse{
		Period:     req.Date,
		Categories: categories,
	}, nil
}

// parsePeriod 解析时间周期
func (s *StatsService) parsePeriod(period, date string) (startDate, endDate time.Time, err error) {
	switch period {
	case "day":
		startDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			return
		}
		endDate = startDate.Add(24*time.Hour - time.Second)

	case "week":
		// 假设date是周的某一天
		t, parseErr := time.Parse("2006-01-02", date)
		if parseErr != nil {
			err = parseErr
			return
		}
		// 计算周一
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		startDate = t.AddDate(0, 0, -weekday+1)
		endDate = startDate.AddDate(0, 0, 7).Add(-time.Second)

	case "month":
		startDate, err = time.Parse("2006-01", date)
		if err != nil {
			return
		}
		endDate = startDate.AddDate(0, 1, 0).Add(-time.Second)

	case "year":
		startDate, err = time.Parse("2006", date)
		if err != nil {
			return
		}
		endDate = startDate.AddDate(1, 0, 0).Add(-time.Second)

	default:
		err = errcode.ErrParams
	}

	return
}
