package domain

type Whitelist struct {
	ID        uint64 `db:"id"`
	Platform  string `db:"platform"`
	UserID    string `db:"user_id"`
	Note      string `db:"note"`
	CreatedAt string `db:"created_at"`
}
