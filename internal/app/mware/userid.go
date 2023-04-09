package mware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
)

func UserID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var userID string

		isCookieOk := true
		usercookie, err := r.Cookie(config.CookieName)
		if err != nil {
			switch {
			default:
				log.Println("Error in cookie")
				w.WriteHeader(http.StatusInternalServerError)
				return
			case errors.Is(err, http.ErrNoCookie):
				isCookieOk = false
			}
		}
		if isCookieOk {
			userIDcrypt := usercookie.Value
			userIDcrypt, err = decodeString(userIDcrypt)
			if err != nil {
				log.Printf("Error in decode: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			userID, isCookieOk = unsignString(userIDcrypt)
		}
		if !isCookieOk {

			userID, err = getNewUserID(16)
			if err != nil {
				log.Printf("Error in rand: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			userIDcrypt := signString(userID)
			userIDcrypt, err = encodeString(userIDcrypt)
			if err != nil {
				log.Printf("Error in encode: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			newcookie := http.Cookie{
				Name:  config.CookieName,
				Value: userIDcrypt,
				Path:  "/",
			}
			http.SetCookie(w, &newcookie)
		}

		ctx := context.WithValue(r.Context(), config.ContextKeyUserID, userID)
		log.Printf("User id: %s", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func getNewUserID(lenbyte int) (string, error) {
	newID := make([]byte, lenbyte)
	_, err := rand.Read(newID)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(newID), nil
}

func decodeString(msg string) (string, error) {
	key := sha256.Sum256([]byte(config.PassCiph))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	encbuf, err := aesgcm.Open(nil, nonce, buf, nil)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(encbuf), nil
}

func encodeString(msg string) (string, error) {
	key := sha256.Sum256([]byte(config.PassCiph))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return "", err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}

	encbuf := aesgcm.Seal(nil, nonce, buf, nil)

	return hex.EncodeToString(encbuf), nil
}

func signString(msg string) string {
	buf, _ := hex.DecodeString(msg)
	key := sha256.Sum256([]byte(config.PassCiph))
	hm := hmac.New(md5.New, key[:])
	hm.Write(buf)
	sign := hex.EncodeToString(hm.Sum(nil))
	return sign + msg
}

func unsignString(msg string) (string, bool) {
	buf, err := hex.DecodeString(msg)
	if err != nil {
		return "", false
	}
	key := sha256.Sum256([]byte(config.PassCiph))
	hm := hmac.New(md5.New, key[:])
	sign := buf[:md5.Size]
	hm.Write(buf[md5.Size:])
	newsign := hm.Sum(nil)
	if hmac.Equal(sign, newsign) {
		return msg[2*md5.Size:], true
	}
	return "", false
}
