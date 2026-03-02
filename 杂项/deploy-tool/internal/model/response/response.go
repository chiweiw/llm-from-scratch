package response

type Base struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Data[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func OK() Base {
	return Base{Code: 0, Message: "ok"}
}

func Fail(message string) Base {
	return Base{Code: 1, Message: message}
}

func OKData[T any](data T) Data[T] {
	return Data[T]{Code: 0, Message: "ok", Data: data}
}

func FailData[T any](message string, zero T) Data[T] {
	return Data[T]{Code: 1, Message: message, Data: zero}
}

