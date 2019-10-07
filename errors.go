package cblcgo

import "fmt"

var (
	ErrNotSupported   = fmt.Errorf("CBL: Not supported")
	ErrNotImplemented = fmt.Errorf("CBL: Not implemented")
	ErrUnknownCommand = fmt.Errorf("CBL: Unknown Command")
	ErrInternalError  = fmt.Errorf("CBL: Internal Error")
	ErrIncorretDatabaseNameFormat = fmt.Errorf("CBL: Incorrect Database Name Format")
	ErrInvalidArguments = fmt.Errorf("CBL: Invalid Arguments")
	ErrProblemClosingDatabase error
	ErrProblemOpeningDatabase error
	ErrProblemPreparingQuery error
	ErrProblemExecutingQuery error
	ErrInvalidCBLType error = fmt.Errorf("CBL: Invalid CBL Type")
	ErrDocumentIsNotReadOnly error = fmt.Errorf("CBL: Document Is Not Read Only")
	ErrDocumentIsReadOnly error = fmt.Errorf("CBL: Document Is Read Only")
	ErrProblemReadingDocument error = fmt.Errorf("CBL: Error Reading Document")
	ErrCBLInternalError error
	ErrProblemSavingDocument error = fmt.Errorf("CBL: Error Saving Document")
	ErrProblemGettingBlobWithData error = fmt.Errorf("CBL: Error Getting Blob With Data.")
	ErrProblemCreatingBlobWithData error = fmt.Errorf("CBL: Error Creating Blob With Data.")
	ErrUnsupportedGoType error = fmt.Errorf("CBL: Unsupported Go type. Use a slice instead.")
)