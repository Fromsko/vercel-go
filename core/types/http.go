package types

type Response struct {
	Code int     `json:"code"`
	Msg  string  `json:"msg"`
	Data any     `json:"data"`
	Err  *string `json:"err,omitempty"`
}

func Success(msg string, data any) *Response {
	return &Response{
		Code: 200,
		Msg:  msg,
		Data: data,
	}
}

func Fail(msg string, err string) *Response {
	return &Response{
		Code: 0,
		Msg:  msg,
		Err:  &err,
	}
}
