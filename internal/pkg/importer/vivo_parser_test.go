package importer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVivoParser_Parse(t *testing.T) {
	parser := &VivoParser{}
	records, err := parser.Parse("test_data/vivo钱包导出.xlsx")

	require.NoError(t, err)
	require.Len(t, records, 4)

	loc, _ := time.LoadLocation("Asia/Shanghai")

	// 第1条记录
	assert.Equal(t, time.Date(2025, 12, 15, 23, 27, 9, 0, loc), records[0].PayTime)
	assert.Equal(t, "38.5", records[0].Amount)
	assert.Equal(t, 1, records[0].BillType) // 支出
	assert.Equal(t, "美团平台商户", records[0].Merchant)
	assert.Equal(t, "餐饮", records[0].CategoryName)
	assert.Equal(t, "vivo钱包", records[0].Platform)

	// 第2条记录
	assert.Equal(t, time.Date(2025, 12, 14, 12, 50, 26, 0, loc), records[1].PayTime)
	assert.Equal(t, "36", records[1].Amount)
	assert.Equal(t, 1, records[1].BillType)
	assert.Equal(t, "福州朴朴电子商务有限公司", records[1].Merchant)
	assert.Equal(t, "烹饪食材", records[1].CategoryName)

	// 第3条记录
	assert.Equal(t, time.Date(2025, 12, 14, 11, 28, 25, 0, loc), records[2].PayTime)
	assert.Equal(t, "244.79", records[2].Amount)
	assert.Equal(t, 1, records[2].BillType)
	assert.Equal(t, "bwh89.net", records[2].Merchant)
	assert.Equal(t, "消费", records[2].CategoryName)

	// 第4条记录
	assert.Equal(t, time.Date(2025, 12, 13, 16, 49, 15, 0, loc), records[3].PayTime)
	assert.Equal(t, "35.86", records[3].Amount)
	assert.Equal(t, 1, records[3].BillType)
	assert.Equal(t, "美团", records[3].Merchant)
	assert.Equal(t, "餐饮", records[3].CategoryName)
}

func TestVivoParser_GetPlatform(t *testing.T) {
	parser := &VivoParser{}
	assert.Equal(t, "vivo钱包", parser.GetPlatform())
}

func TestVivoParser_Parse_FileNotFound(t *testing.T) {
	parser := &VivoParser{}
	_, err := parser.Parse("test_data/不存在的文件.xlsx")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "excel解析打开文件失败")
}
