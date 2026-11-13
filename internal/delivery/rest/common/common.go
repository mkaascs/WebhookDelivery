package common

import (
	apiv1 "webhook-delivery/gen/rest/v1"
)

const (
	defaultPage  = 1
	defaultLimit = 10
)

func InternalError() apiv1.Error {
	return apiv1.Error{Code: new("internal_server_error"), Message: "internal server error"}
}

func InvalidURLError() apiv1.Error {
	return apiv1.Error{Code: new("invalid_url"), Message: "invalid url"}
}

func Pagination(page, limit *int) (int, int) {
	p, l := defaultPage, defaultLimit
	if page != nil {
		p = *page
	}

	if limit != nil {
		l = *limit
	}

	return p, l
}

func GetDefaultArrayIfNull(types []string) []string {
	if types == nil {
		return []string{}
	}

	return types
}
