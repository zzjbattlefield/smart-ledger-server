package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/repository"
	"smart-ledger-server/pkg/errcode"
)

// BillService 账单服务
type BillService struct {
	billRepo     BillRepo
	categoryRepo CategoryRepo
}

// NewBillService 创建账单服务
func NewBillService(billRepo BillRepo, categoryRepo CategoryRepo) *BillService {
	return &BillService{
		billRepo:     billRepo,
		categoryRepo: categoryRepo,
	}
}

// Create 创建账单
func (s *BillService) Create(ctx context.Context, userID uint64, req *dto.CreateBillRequest) (*dto.BillResponse, error) {
	bill := &model.Bill{
		UUID:       uuid.New().String(),
		UserID:     userID,
		Amount:     req.Amount,
		BillType:   model.BillType(req.BillType),
		Platform:   req.Platform,
		Merchant:   req.Merchant,
		CategoryID: req.CategoryID,
		PayTime:    req.PayTime,
		PayMethod:  req.PayMethod,
		OrderNo:    req.OrderNo,
		Remark:     req.Remark,
	}

	// 转换账单明细
	var items []model.BillItem
	if len(req.Items) > 0 {
		items = make([]model.BillItem, len(req.Items))
		for i, item := range req.Items {
			items[i] = model.BillItem{
				Name:     item.Name,
				Price:    item.Price,
				Quantity: item.Quantity,
			}
		}
	}

	if err := s.billRepo.CreateWithItems(ctx, bill, items); err != nil {
		return nil, errcode.ErrBillCreateFailed
	}

	return s.GetByID(ctx, userID, bill.ID)
}

// CreateFromAI 从AI识别结果创建账单
func (s *BillService) CreateFromAI(ctx context.Context, userID uint64, aiResult *dto.AIRecognizeResponse, imagePath string) (*dto.BillResponse, error) {
	// 查找分类
	var categoryID *uint64
	if aiResult.SubCategory != "" {
		category, err := s.categoryRepo.GetByName(ctx, aiResult.SubCategory)
		if err == nil {
			categoryID = &category.ID
		}
	}
	if categoryID == nil && aiResult.Category != "" {
		category, err := s.categoryRepo.GetByName(ctx, aiResult.Category)
		if err == nil {
			categoryID = &category.ID
		}
	}

	bill := &model.Bill{
		UUID:        uuid.New().String(),
		UserID:      userID,
		Amount:      aiResult.Amount,
		BillType:    model.BillTypeExpense,
		Platform:    aiResult.Platform,
		Merchant:    aiResult.Merchant,
		CategoryID:  categoryID,
		PayTime:     aiResult.PayTime,
		PayMethod:   aiResult.PayMethod,
		OrderNo:     aiResult.OrderNo,
		ImagePath:   imagePath,
		Confidence:  aiResult.Confidence,
		IsConfirmed: false,
	}

	// 转换账单明细
	var items []model.BillItem
	if len(aiResult.Items) > 0 {
		items = make([]model.BillItem, len(aiResult.Items))
		for i, item := range aiResult.Items {
			items[i] = model.BillItem{
				Name:     item.Name,
				Price:    item.Price,
				Quantity: item.Quantity,
			}
		}
	}

	if err := s.billRepo.CreateWithItems(ctx, bill, items); err != nil {
		return nil, errcode.ErrBillCreateFailed
	}

	return s.GetByID(ctx, userID, bill.ID)
}

// GetByID 获取账单详情
func (s *BillService) GetByID(ctx context.Context, userID, id uint64) (*dto.BillResponse, error) {
	bill, err := s.billRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrBillNotFound
		}
		return nil, errcode.ErrServer
	}

	// 检查权限
	if bill.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	return s.toBillResponse(bill), nil
}

// List 获取账单列表
func (s *BillService) List(ctx context.Context, userID uint64, req *dto.BillListRequest) (*dto.BillListResponse, error) {
	req.SetDefaults()

	query := &repository.BillQuery{
		UserID:   userID,
		Page:     req.Page,
		PageSize: req.PageSize,
		Keyword:  req.Keyword,
	}

	// 解析日期
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			query.StartDate = &t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			// 结束日期设为当天23:59:59
			endOfDay := t.Add(24*time.Hour - time.Second)
			query.EndDate = &endOfDay
		}
	}

	if req.CategoryID > 0 {
		query.CategoryID = &req.CategoryID
	}
	if req.BillType > 0 {
		query.BillType = &req.BillType
	}

	bills, total, err := s.billRepo.List(ctx, query)
	if err != nil {
		return nil, errcode.ErrServer
	}

	list := make([]dto.BillResponse, len(bills))
	for i, bill := range bills {
		list[i] = *s.toBillResponse(&bill)
	}

	return &dto.BillListResponse{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		List:     list,
	}, nil
}

// Update 更新账单
func (s *BillService) Update(ctx context.Context, userID, id uint64, req *dto.UpdateBillRequest) (*dto.BillResponse, error) {
	bill, err := s.billRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrBillNotFound
		}
		return nil, errcode.ErrServer
	}

	// 检查权限
	if bill.UserID != userID {
		return nil, errcode.ErrForbidden
	}

	// 更新字段
	if !req.Amount.IsZero() {
		bill.Amount = req.Amount
	}
	if req.BillType > 0 {
		bill.BillType = model.BillType(req.BillType)
	}
	if req.Platform != "" {
		bill.Platform = req.Platform
	}
	if req.Merchant != "" {
		bill.Merchant = req.Merchant
	}
	if req.CategoryID != nil {
		bill.CategoryID = req.CategoryID
	}
	if req.PayTime != nil {
		bill.PayTime = *req.PayTime
	}
	if req.PayMethod != "" {
		bill.PayMethod = req.PayMethod
	}
	if req.OrderNo != "" {
		bill.OrderNo = req.OrderNo
	}
	if req.Remark != "" {
		bill.Remark = req.Remark
	}
	if req.IsConfirmed != nil {
		bill.IsConfirmed = *req.IsConfirmed
	}

	if err := s.billRepo.Update(ctx, bill); err != nil {
		return nil, errcode.ErrBillUpdateFailed
	}

	return s.GetByID(ctx, userID, id)
}

// Delete 删除账单
func (s *BillService) Delete(ctx context.Context, userID, id uint64) error {
	bill, err := s.billRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrBillNotFound
		}
		return errcode.ErrServer
	}

	// 检查权限
	if bill.UserID != userID {
		return errcode.ErrForbidden
	}

	if err := s.billRepo.Delete(ctx, id); err != nil {
		return errcode.ErrBillDeleteFailed
	}

	return nil
}

// toBillResponse 转换为账单响应
func (s *BillService) toBillResponse(bill *model.Bill) *dto.BillResponse {
	resp := &dto.BillResponse{
		ID:          bill.ID,
		UUID:        bill.UUID,
		Amount:      bill.Amount,
		BillType:    int(bill.BillType),
		Platform:    bill.Platform,
		Merchant:    bill.Merchant,
		PayTime:     bill.PayTime,
		PayMethod:   bill.PayMethod,
		OrderNo:     bill.OrderNo,
		Remark:      bill.Remark,
		Confidence:  bill.Confidence,
		IsConfirmed: bill.IsConfirmed,
		CreatedAt:   bill.CreatedAt,
	}

	if bill.Category != nil {
		resp.Category = &dto.CategoryResponse{
			ID:   bill.Category.ID,
			Name: bill.Category.Name,
			Icon: bill.Category.Icon,
		}
	}

	if len(bill.Items) > 0 {
		resp.Items = make([]dto.BillItemResponse, len(bill.Items))
		for i, item := range bill.Items {
			resp.Items[i] = dto.BillItemResponse{
				ID:       item.ID,
				Name:     item.Name,
				Price:    item.Price,
				Quantity: item.Quantity,
			}
		}
	}

	return resp
}
