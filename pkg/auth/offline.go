package auth

import (
	"crypto/md5"
	"fmt"
)

type OfflineAuth struct {
	Username string
}

func NewOfflineAuth(username string) Auth {
	return &OfflineAuth{
		Username: username,
	}
}

func (a *OfflineAuth) GetAuthData() (string, string) {
	return a.Username, genUUID(a.Username)
}

func genUUID(username string) string {
	hash := md5.Sum([]byte("OfflinePlayer:" + username))
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:16])
}
