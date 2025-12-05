package entity

type UserThread struct {
	UserId   int64 `db:"user_id"`
	ThreadId int64 `db:"thread_id"`
}
