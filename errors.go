package vcard

import (
	"errors"
	"fmt"
)

// Signifies any error from vCard library
var ErrVCard = errors.New("vCard")

// Signifies decoding was unsuccessful because of the syntax of the vCard document and not
// more significant errors.
var ErrParsing = fmt.Errorf("%w: parsing error in Decoder", ErrVCard)

// Signifies decoding was successful but there are more tokens left.
// This could be the case when trying to decode a document of multiple vCards into a single struct or a map.
var ErrLeftoverTokens = fmt.Errorf("%w: leftover tokens", ErrParsing)
