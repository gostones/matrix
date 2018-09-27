//based on https://github.com/kr/session/blob/master/session.go
package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/nacl/secretbox"
	"log"
	"net/http"
	"time"
)

const secret = "54686520776f7264206875736b79206f726967696e617465642066726f6d2074686520776f726420726566657272696e6720746f204172637469632070656f706c6520696e2067656e6572616c2c20496e7569742028612e6b2e612e2045736b696d6f73292c202e2e2e6b6e6f776e20617320276875736b69657327"
const cookieKey = "matrix_session"

var cookieConfig *Config

type MatrixSession struct {
	Key       string `json:"key"`
	Port      int    `json:"port"`
	Timestamp int64  `json:"timestamp"`
}

func SetCookie(w http.ResponseWriter, v *MatrixSession) error {
	return setCookie(w, v, cookieConfig)
}

func GetCookie(r *http.Request, v *MatrixSession) error {
	return getCookie(r, v, cookieConfig)
}

func ClearCookie(name string, w http.ResponseWriter) {
	expired := &http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Now().Add(-100 * time.Hour),
		MaxAge:  -1,
	}
	http.SetCookie(w, expired)
}

var secretKeys []*[32]byte

//
func init() {

	var sk [32]byte
	secretKeyBytes, _ := hex.DecodeString(secret)
	copy(sk[:], secretKeyBytes)
	secretKeys = []*[32]byte{&sk}

	cookieConfig = CookieConfig("/")
}

func CookieConfig(path string) *Config {
	return &Config{Name: cookieKey, Path: path, HTTPOnly: true, Secure: false, Keys: secretKeys}
}

//
const (
	maxSize = 4093
)

type Config struct {
	// The cookie name.
	// If empty, "session" is used.
	Name string

	// The cookie path.
	// If empty, "/" is used.
	Path string

	// The cookie domain.
	// If empty, the request host is used.
	Domain string

	// Whether the cookie should be limited to HTTPS.
	Secure bool

	// Whether the cookie will not be available to JavaScript.
	HTTPOnly bool

	// Maximum idle time for a session.
	// This is used to set cookie expiration and
	// enforce a TTL on secret boxes.
	// If 0, it is taken to be 100 years.
	MaxAge time.Duration

	// List of acceptable secretbox keys for decoding stored sessions.
	// Element 0 will be used for encoding.
	// See golang.org/x/crypto/nacl/secretbox.
	Keys []*[32]byte
}

func (c *Config) maxAge() time.Duration {
	if c.MaxAge == 0 {
		return 100 * 365 * 24 * time.Hour
	}
	return c.MaxAge
}

func (c *Config) name() string {
	if c.Name == "" {
		return "session"
	}
	return c.Name
}

// Indicates the encoded session cookie is too long
// to expect web browsers to store it.
var (
	ErrTooLong = errors.New("encoded session too long")
	ErrInvalid = errors.New("invalid session cookie")
)

// Get decodes a session from r into v.
// See encoding/json for decoding behavior.
func getCookie(r *http.Request, v interface{}, config *Config) error {
	cookie, err := r.Cookie(config.name())
	if err != nil {
		return err
	}
	return decodeCookie(cookie, v, config)
}

func decodeCookie(cookie *http.Cookie, v interface{}, config *Config) error {
	t, err := base64.URLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return err
	}
	var nonce [24]byte
	copy(nonce[:], t)
	for _, key := range config.Keys {
		if tb, ok := secretbox.Open(nil, t[24:], &nonce, key); ok {
			ts := binary.BigEndian.Uint64(tb)
			if time.Since(time.Unix(int64(ts), 0)) > config.maxAge() {
				return ErrInvalid
			}
			b := tb[8:]
			return json.Unmarshal(b, v)
		}
	}
	return ErrInvalid
}

// Set encodes a session from v into a cookie on w.
// See encoding/json for encoding behavior.
func setCookie(w http.ResponseWriter, v interface{}, config *Config) error {
	s := encodeCookie(v, config)
	if s == "" {
		s = "ERR"
	}
	w.Header().Add("Set-Cookie", s)
	return nil
}

func encodeCookie(v interface{}, config *Config) string {
	now := time.Now()
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("encodeCookie: marshal %v %v", v, err)
		return ""
	}
	tb := make([]byte, len(b)+8)
	binary.BigEndian.PutUint64(tb, uint64(now.Unix()))
	copy(tb[8:], b)
	var nonce [24]byte
	_, err = rand.Read(nonce[:])
	if err != nil {
		log.Printf("encodeCookie: nonce %v %v", v, err)
		return ""
	}
	out := secretbox.Seal(nonce[:], tb, &nonce, config.Keys[0])

	cookie := &http.Cookie{
		Name:  config.name(),
		Value: base64.URLEncoding.EncodeToString(out),
		//Expires:  now.Add(config.maxAge()),
		Path:     config.Path,
		Domain:   config.Domain,
		Secure:   config.Secure,
		HttpOnly: config.HTTPOnly,
	}
	if cookie.Path == "" {
		cookie.Path = "/"
	}
	s := cookie.String()
	if len(s) > maxSize {
		log.Printf("encodeCookie: max size exeeded %v %v", maxSize, s)
		return ""
	}

	return s
}
