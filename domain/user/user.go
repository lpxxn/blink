package user

const (
	StatusInactive = 0
	StatusActive   = 1
	StatusBanned   = 2
)

type User struct {
	SnowflakeID  int64
	Email        string
	Name         string
	WechatID     string
	Phone        string
	PasswordHash string
	PasswordSalt string
	Status       int
	Role         string
}
