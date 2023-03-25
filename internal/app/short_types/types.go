package types

import "errors"

type ShortURL struct {
	Short string
	URL   string
}

const LenShortURL = 5
const OurHost = "http://localhost:8080/"
const ReturnShortWithHost = true

var (
	ErrNoSuchRecord   = errors.New("no such record")
	ErrInvalidReqBody = errors.New("invalid request body")
	ErrEmptyReqBody   = errors.New("empty request body")
	ErrURLNotCorrect  = errors.New("given url is not correct")
	ErrNoFreeIDs      = errors.New("no free short url")
)
