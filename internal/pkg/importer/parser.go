package importer

import "time"

type BillRecord struct {
	PayTime      time.Time
	Amount       string
	BillType     int    //1=支出, 2=收入
	Merchant     string // 商户/备注
	CategoryName string //分类名称
	Platform     string //平台
}

type ExcelParser interface {
	Parse(filepath string) ([]BillRecord, error)
	GetPlatform() string
}
