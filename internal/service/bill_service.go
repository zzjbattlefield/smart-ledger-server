package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	pkgerrors "github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"smart-ledger-server/internal/model"
	"smart-ledger-server/internal/model/dto"
	"smart-ledger-server/internal/pkg/importer"
	"smart-ledger-server/internal/pkg/logger"
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
	// 校验分类归属（用户级分类）
	var categoryID *uint64
	if req.CategoryID != nil && *req.CategoryID != 0 {
		category, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrCategoryNotFound
			}
			return nil, errcode.ErrServer
		}
		if category.UserID != userID {
			return nil, errcode.ErrCategoryNotFound
		}
		categoryID = req.CategoryID
	}

	bill := &model.Bill{
		UUID:       uuid.New().String(),
		UserID:     userID,
		Amount:     req.Amount,
		BillType:   model.BillType(req.BillType),
		Platform:   req.Platform,
		Merchant:   req.Merchant,
		CategoryID: categoryID,
		PayTime:    req.PayTime,
		PayMethod:  req.PayMethod,
		OrderNo:    req.OrderNo,
		Remark:     req.Remark,
	}

	if err := s.billRepo.Create(ctx, bill); err != nil {
		return nil, errcode.ErrBillCreateFailed
	}

	return s.GetByID(ctx, userID, bill.ID)
}

// CreateFromAI 从AI识别结果创建账单
func (s *BillService) CreateFromAI(ctx context.Context, userID uint64, aiResult *dto.AIRecognizeResponse, imagePath string) (*dto.BillResponse, error) {
	// 根据 AI 返回的 bill_type 确定账单类型和分类类型
	billType := model.BillTypeExpense
	categoryType := model.CategoryTypeExpense
	if aiResult.BillType == 2 {
		billType = model.BillTypeIncome
		categoryType = model.CategoryTypeIncome
	}

	// 查找分类（按类型过滤）
	var categoryID *uint64
	if aiResult.SubCategory != "" {
		category, err := s.categoryRepo.GetByNameAndType(ctx, userID, aiResult.SubCategory, categoryType)
		if err == nil {
			categoryID = &category.ID
		}
	}
	if categoryID == nil && aiResult.Category != "" {
		category, err := s.categoryRepo.GetByNameAndType(ctx, userID, aiResult.Category, categoryType)
		if err == nil {
			categoryID = &category.ID
		}
	}
	payTime := time.Now()
	if aiResult.PayTime != "" {
		local, _ := time.LoadLocation("Asia/Shanghai")
		if parseTime, err := time.ParseInLocation(time.RFC3339, aiResult.PayTime, local); err == nil {
			payTime = parseTime
		} else {
			logger.Log.Info("解析账单支付时间出现错误", zap.String("aiResult下的paytime", aiResult.PayTime), zap.Error(err))
		}
	}

	bill := &model.Bill{
		UUID:        uuid.New().String(),
		UserID:      userID,
		Amount:      aiResult.Amount,
		BillType:    billType,
		Platform:    aiResult.Platform,
		Merchant:    aiResult.Merchant,
		CategoryID:  categoryID,
		PayTime:     payTime,
		PayMethod:   aiResult.PayMethod,
		OrderNo:     aiResult.OrderNo,
		ImagePath:   imagePath,
		Confidence:  aiResult.Confidence,
		IsConfirmed: false,
	}

	if err := s.billRepo.Create(ctx, bill); err != nil {
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
		if *req.CategoryID == 0 {
			bill.CategoryID = nil
		} else {
			category, err := s.categoryRepo.GetByID(ctx, *req.CategoryID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil, errcode.ErrCategoryNotFound
				}
				return nil, errcode.ErrServer
			}
			if category.UserID != userID {
				return nil, errcode.ErrCategoryNotFound
			}
			bill.CategoryID = req.CategoryID
			bill.Category = nil
		}
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
		// 防御性：避免账单错误关联到其他用户的分类
		if bill.Category.UserID == bill.UserID {
			resp.Category = &dto.CategoryResponse{
				ID:   bill.Category.ID,
				Name: bill.Category.Name,
				Type: int(bill.Category.Type),
				Icon: bill.Category.Icon,
			}
		}
	}

	return resp
}

func (s *BillService) ImportFromExcel(ctx context.Context, userID uint64, filePath, parserType string) (response *dto.BillImportResponse, err error) {
	response = &dto.BillImportResponse{}
	// 1. 创建解析器
	parser, _ := importer.NewParser(importer.ParserTypeVivo)
	parseResult, err := parser.Parse(filePath)
	if err != nil {
		return nil, err
	}
	//获取解析成功的数据
	records := parseResult.Records
	//处理解析阶段就失败的数据
	for _, parseError := range parseResult.Errors {
		response.Failed++
		response.Errors = append(response.Errors, dto.ImportError{
			Row:     parseError.Row,
			Column:  parseError.Column,
			Message: parseError.Message,
			RowData: parseError.RowData,
		})
	}
	//获取当前用户的所有分类
	categorys, err := s.categoryRepo.GetAll(ctx, userID)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "无法获取当前用户的分类列表")
	}
	categoryMap := make(map[string]uint64, len(categorys)-1)
	//将分类转成map
	for _, category := range categorys {
		categoryMap[category.Name] = category.ID
	}
	otherCategoryID, hasOtherCategory := categoryMap["未分类"] //判断是否存在未分类这个类目
	for _, reocrd := range records {
		rowNum := reocrd.Row
		amount, err := decimal.NewFromString(reocrd.Amount)
		if err != nil {
			response.Failed++
			response.Errors = append(response.Errors, dto.ImportError{
				Row:     rowNum,
				Message: "金额格式转换错误",
				Column:  "金额",
				RowData: reocrd.RowData,
			})
			continue
		}
		//确认分类
		var categoryID uint64
		if catID, exists := categoryMap[reocrd.CategoryName]; !exists {
			//当前分类不存在，放到未分类的类目，如果没有未分类则需要帮他创建一个
			if !hasOtherCategory {
				otherCategory := &model.Category{
					Name:     "未分类",
					UserID:   userID,
					ParentID: 0,
				}
				s.categoryRepo.Create(ctx, otherCategory)
				otherCategoryID = otherCategory.ID
				hasOtherCategory = true
			}
			categoryID = otherCategoryID
		} else {
			categoryID = catID
		}
		BillType := model.BillTypeExpense
		if reocrd.BillType == 2 {
			BillType = model.BillTypeIncome
		}
		//创建账单
		bill := &model.Bill{
			UUID:       uuid.New().String(),
			UserID:     userID,
			Amount:     amount,
			BillType:   BillType,
			CategoryID: &categoryID,
			PayTime:    reocrd.PayTime,
			Merchant:   reocrd.Merchant,
		}
		if err := s.billRepo.Create(ctx, bill); err != nil {
			response.Failed++
			response.Errors = append(response.Errors, dto.ImportError{
				Row:     rowNum,
				Message: "创建账单失败",
			})
			continue
		}
		response.Total++
	}
	return response, nil
}
