package importer

import "time"

type BillRecord struct {
	Row          int
	PayTime      time.Time
	Amount       string
	BillType     int    //1=支出, 2=收入
	Merchant     string // 商户/备注
	CategoryName string //分类名称
	Platform     string //平台
	RowData      map[string]string
}

type ParseResult struct {
	Records []BillRecord
	Errors  []ParseError
}

type ParseError struct {
	Row     int
	Column  string
	Message string
	RowData map[string]string
}

type ExcelParser interface {
	Parse(filepath string) (*ParseResult, error)
	GetPlatform() string
}
