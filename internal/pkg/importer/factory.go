package importer

import "fmt"

type ParserType string

const (
	ParserTypeVivo ParserType = "vivo"
	ParserTypeWX   ParserType = "wx"  //TODO: 微信账单导入
	ParserTypeAli  ParserType = "ali" //TODO: 支付宝账单导入
)

func NewParser(parserType ParserType) (ExcelParser, error) {
	switch parserType {
	case ParserTypeVivo:
		return NewVivoParser(), nil
	default:
		return nil, fmt.Errorf("暂不支持的解析器类型: %s", parserType)
	}
}
