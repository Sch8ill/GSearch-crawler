package types

type Site struct {
	Url          string
	Host         string
	Scheme       string
	Text         []string
	Links        []string
	Timestamp    string
	Depth        int
	FoundThrough string
	Type         string
	Err          error
}
