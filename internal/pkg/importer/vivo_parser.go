package importer

import (
	"smart-ledger-server/internal/pkg/logger"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type VivoParser struct{}

func NewVivoParser() *VivoParser {
	return &VivoParser{}
}

func (p *VivoParser) Parse(filepath string) ([]BillRecord, error) {
	var records []BillRecord
	excelFile, err := excelize.OpenFile(filepath)
	if err != nil {
		err = pkgerrors.Wrap(err, "excel解析打开文件失败")
		return records, err
	}
	defer func() {
		if err = excelFile.Close(); err != nil {
			logger.Log.Error("excel解析关闭文件失败", zap.Error(err))
		}
	}()
	shellName := excelFile.GetSheetName(0)
	rows, err := excelFile.GetRows(shellName)
	if err != nil {
		err = pkgerrors.Wrap(err, "获取excel行数据失败")
		return records, err
	}
	if len(rows) < 2 {
		return nil, pkgerrors.New("文件为空或只有表头")
	}
	for i, row := range rows[1:] {
		if len(row) < 6 {
			//跳过不完整的行
			continue
		}
		PayTime, err := parseTime(row[0])
		if err != nil {
			return records, pkgerrors.Wrapf(err, "第 %d 行时间解析失败", i+2)
		}
		BillType := 1
		if row[3] == "收入" {
			BillType = 2
		}
		records = append(records, BillRecord{
			PayTime:      PayTime,
			Amount:       row[5], // 金额
			BillType:     BillType,
			Merchant:     row[4], // 备注作为商户
			CategoryName: row[2], // 记账分类
			Platform:     p.GetPlatform(),
		})
	}
	return records, nil
}

func (p *VivoParser) GetPlatform() string {
	return "vivo钱包"
}

// parseTime 将vivo钱包导入的时间转成time.Time格式
func parseTime(s string) (time.Time, error) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return time.ParseInLocation("2006-01-02 15:04:05", s, location)
}
