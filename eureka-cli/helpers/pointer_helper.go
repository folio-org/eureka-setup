package helpers

func StringP(value string) *string {
	return &value
}

func BoolP(value bool) *bool {
	return &value
}

func IntP(value int) *int {
	return &value
}
