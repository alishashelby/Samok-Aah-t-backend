package entity

import "log"

const (
	LimitMaxValue     = 40
	LimitDefaultValue = 10
)

const (
	ParamPage  = "page"
	ParamLimit = "limit"
)

type Options struct {
	Page   int64
	Limit  int64
	Offset int64
}

func NewOptions(page, limit int64) *Options {
	if page < 1 {
		log.Printf("pagination page < 1, now page = 1")
		page = 1
	}

	if limit < 1 {
		log.Printf("pagination limit < 1, now limit = %d", LimitDefaultValue)
		limit = LimitDefaultValue
	}

	if limit > LimitMaxValue {
		log.Printf("pagination limit > max value = %d", LimitMaxValue)
		limit = LimitMaxValue
	}

	return &Options{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}
