package types

import "errors"

type ShortURL struct {
	Short string
	URL   string
}

const LenShortURL = 10
const OurHost = "http://localhost:8080/"
const ReturnShortWithHost = true

var ErrNoSuchRecord = errors.New("no such record")
var ErrInvalidReqBody = errors.New("invalid request body")
var ErrEmptyReqBody = errors.New("empty request body")
var ErrURLNotCorrect = errors.New("given url is not correct")
