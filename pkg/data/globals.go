package data

var (
	caseInsensitive bool = false
)

func SetCaseInsensitive() {
	caseInsensitive = true
}

func GetCaseInsensitive() bool {
	return caseInsensitive
}
