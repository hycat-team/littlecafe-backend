package jwttype

type JwtType string

const (
	AccessToken        JwtType = "ACCESS_TOKEN"
	RefreshToken       JwtType = "REFRESH_TOKEN"
	ResetPasswordToken JwtType = "RESET_PASSWORD_TOKEN"
)
