package structure

import (
	"database/sql/driver"
	"errors"
)

type BitBool bool

// 实现 sql.Scanner 接口
func (b *BitBool) Scan(value any) error {
	if value == nil {
		*b = false
		return nil
	}
	// MySQL 返回的是 []uint8 类型（如 0x00 或 0x01）
	if v, ok := value.([]uint8); ok {
		*b = v[0] == 1
		return nil
	}
	return errors.New("无法转换值到 BitBool")
}

// 实现 driver.Valuer 接口
func (b BitBool) Value() (driver.Value, error) {
	if b {
		return []byte{1}, nil
	}
	return []byte{0}, nil
}
