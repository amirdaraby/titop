package config

type System struct {
	ClkTck int64
	CoresCount int64
	PageSize int64
	TotalMem int64 // in RSS
}
