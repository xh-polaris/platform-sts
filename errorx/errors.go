package errorx

import "google.golang.org/grpc/status"

var (
	ErrCannotDeleteObject = status.Error(10001, "can not delete object")
)
