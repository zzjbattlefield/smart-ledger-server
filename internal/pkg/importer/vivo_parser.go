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

// vivo账单的列名
var vivoColumns = []string{"交易时间", "交易单号", "记账分类", "收支类型", "备注", "交易金额"}

func (p *VivoParser) Parse(filepath string) (*ParseResult, error) {
	var records []BillRecord
	var parseErrors []ParseError
	excelFile, err := excelize.OpenFile(filepath)
	if err != nil {
		err = pkgerrors.Wrap(err, "excel解析打开文件失败")
		return nil, err
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
		return nil, err
	}
	if len(rows) < 2 {
		return nil, pkgerrors.New("文件为空或只有表头")
	}
	for i, row := range rows[1:] {
		if len(row) != 6 {
			//跳过不完整的行
			continue
		}
		PayTime, err := parseTime(row[0])
		if err != nil {
			parseErrors = append(parseErrors, ParseError{
				Row:     i + 2,
				Column:  "账单日期",
				Message: "账单日期格式错误",
				RowData: p.buildRawData(row),
			})
			continue
		}
		BillType := 1
		if row[3] == "收入" {
			BillType = 2
		}
		records = append(records, BillRecord{
			RowData:      p.buildRawData(row),
			Row:          i + 2,
			PayTime:      PayTime,
			Amount:       row[5], // 金额
			BillType:     BillType,
			Merchant:     row[4], // 备注作为商户
			CategoryName: row[2], // 记账分类
			Platform:     p.GetPlatform(),
		})
	}
	return &ParseResult{
		Records: records,
		Errors:  parseErrors,
	}, nil
}

func (p *VivoParser) GetPlatform() string {
	return "vivo钱包"
}

// parseTime 将vivo钱包导入的时间转成time.Time格式
func parseTime(s string) (time.Time, error) {
	location, _ := time.LoadLocation("Asia/Shanghai")
	return time.ParseInLocation("2006-01-02 15:04:05", s, location)
}

func (p *VivoParser) buildRawData(rows []string) map[string]string {
	rowsData := make(map[string]string)
	for i, row := range rows {
		rowsData[vivoColumns[i]] = row
	}
	return rowsData
}
