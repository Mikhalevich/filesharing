package httpcode

type Error interface {
	StatusCode() int
	Description() string
}
