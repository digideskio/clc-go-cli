package db

type ActionLogEntry struct {
	Timestamp int64
	Message   string
	Details   string
	User      string
}

type Promotion struct {
	Id   int64
	Code string
}