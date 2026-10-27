package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"

	"homework-station/internal/model"
)

func (svc *UserService) BindQQ(u model.User, code string) (model.User, error) {
	if u.QQ != nil {
		return model.User{}, Reject(409, "已绑定 QQ，请联系管理员解绑")
	}
	t, e := jwt.Parse(code, func(t *jwt.Token) (interface{}, error) { return []byte(svc.settings.Get("binding_secret")), nil }, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired(), jwt.WithIssuedAt(), jwt.WithAudience("homework-qq-bind"))
	if e != nil || !t.Valid {
		return model.User{}, Reject(400, "绑定码无效或已过期")
	}
	claims := t.Claims.(jwt.MapClaims)
	qq, _ := claims["qq"].(string)
	exp, e1 := claims.GetExpirationTime()
	iat, e2 := claims.GetIssuedAt()
	ttl, _ := claims["ttl"].(float64)
	if !regexp.MustCompile(`^[1-9][0-9]{4,14}$`).MatchString(qq) || e1 != nil || e2 != nil || iat == nil || exp == nil || ttl != 300 || exp.Unix()-iat.Unix() != 300 {
		return model.User{}, Reject(400, "绑定码格式不正确")
	}
	hash := sha256.Sum256([]byte(code))
	e = svc.db.Transaction(func(tx *gorm.DB) error {
		if e := tx.Create(&model.UsedCode{ID: hex.EncodeToString(hash[:]), ExpiresAt: exp.Time}).Error; e != nil {
			return e
		}
		resultDB := tx.Model(&model.User{}).Where("id = ? AND qq IS NULL", u.ID).Update("qq", qq)
		if resultDB.Error != nil {
			return resultDB.Error
		}
		if resultDB.RowsAffected != 1 {
			return fmt.Errorf("绑定状态已改变")
		}
		return nil
	})
	if e != nil {
		return model.User{}, Reject(409, "绑定码已使用或 QQ 已被绑定")
	}
	svc.db.Where("expires_at < ?", time.Now()).Delete(&model.UsedCode{})
	svc.db.First(&u, u.ID)
	return u, nil
}
