package httperror

type Code int

const (
	CodeInternalError Code = 1
	CodeInvalidParams Code = 2
	CodeUnauthorized  Code = 3
	CodeAlreadyExist  Code = 4
	CodeNotExist      Code = 5
	CodeNotMatch      Code = 6
)

func (c Code) Int() int {
	return int(c)
}
