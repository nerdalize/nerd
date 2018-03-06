package populator

// ErrNoSuchPopulator implements the error interface
type ErrNoSuchPopulator string

func (e ErrNoSuchPopulator) Error() string { return string(e) }
