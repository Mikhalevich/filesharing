package httpcode

type Code int

const (
	CodeInternalError Code = 1
	CodeInvalidParams Code = 2
	CodeUnauthorized  Code = 3
	CodeAlreadyExist  Code = 600
	CodeNotExist      Code = 601
	CodeNotMatch      Code = 602
)

func (c Code) Int() int {
	return int(c)
}
