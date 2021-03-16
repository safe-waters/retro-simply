package store

type (
	DataAlreadyExistsError struct{ Err error }
	DataDoesNotExistError  struct{ Err error }
)

func (d DataAlreadyExistsError) Error() string { return d.Err.Error() }

func (d DataDoesNotExistError) Error() string { return d.Err.Error() }
