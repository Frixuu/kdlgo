package kdlgo

import (
	"errors"
	"strconv"
)

type KDLErrorType string

const (
	KDLEmptyArray     = "Array is empty"
	KDLDifferentKey   = "All keys of KDLObject to convert to document should be the same"
	KDLInvalidKeyChar = "Invalid character for key"
	KDLInvalidSyntax  = "Invalid syntax"
	KDLInvalidType    = "Invalid KDLType"

	// These should be caught and handled internally
	kdlKeyOnly  = "Key only"
	kdlEndOfObj = "End of KDLObject"
)

func wrapError(err error) error {
	return errors.New(err.Error() + "\nOn line " + strconv.Itoa(line) + " column " + strconv.Itoa(pos))
}

func differentKeysErr() error {
	return errors.New(KDLDifferentKey)
}

func emptyArrayErr() error {
	return errors.New(KDLEmptyArray)
}

func invalidKeyCharErr() error {
	return errors.New(KDLInvalidKeyChar)
}

func invalidSyntaxErr() error {
	return errors.New(KDLInvalidSyntax)
}

func invalidTypeErr() error {
	return errors.New(KDLInvalidType)
}

func keyOnlyErr() error {
	return errors.New(kdlKeyOnly)
}

func endOfObjErr() error {
	return errors.New(kdlEndOfObj)
}