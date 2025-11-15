package errx

import "fmt"

var ErrInvalid = New(BadRequest, "Invalid JSON format.")
var ErrCaptcha = New(BadRequest, "We could not confirm you are human")
var ErrOptions = New(BadRequest, "You have to choose 1 option at least to generate.")

func ErrGeneration(err error) *Error {
	return New(BadRequest, fmt.Sprintf("We were unable to generate the response: %s", err.Error()))
}
