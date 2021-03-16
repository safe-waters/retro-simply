package data

type (
	PasswordInvalidError struct{ Err error }
	RoomIdInvalidError   struct{ Err error }
)

func (p PasswordInvalidError) Error() string { return p.Err.Error() }

func (r RoomIdInvalidError) Error() string { return r.Err.Error() }
