package commerrs

// 占位用 todo
var ErrTODO = &APIError{100000, "TODO"}

// 鉴权失败
var ErrUnauthorized = &APIError{100401, "unauthorized"}

// id格式不正确 (本项目中一般指 mongodb 的 ObjectID)
var ErrInvalidObjectID = &APIError{100001, "invalid ID format"}

// 没有找到数据
var ErrDataNotFound = &APIError{100002, "requested data not found"}

// 服务错误 (用作兜底)
var ErrServiceError = &APIError{100999, "service error"}
