package supermarket

type ServiceUnavailableError struct {
	error
}

func (e ServiceUnavailableError) Error() string {
	return "service unavailable"
}
