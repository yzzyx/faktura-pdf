package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/yzzyx/zerr"
	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrInvalidHash              = errors.New("invalid hash")
	ErrUnknownPasswordAlgorithm = errors.New("invalid password algorithm")
)

type User struct {
	ID       int
	Username string
	Name     string
	Email    string
	Company  Company

	password string
}

type UserFilter struct {
	ID       int
	Username string
}

// GenerateRandomString generates a random string of the specified length containing a-z, A-Z, 0-9
func GenerateRandomString(length int) ([]byte, error) {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	rnd := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, rnd)
	if err != nil {
		return []byte{}, zerr.Wrap(err)
	}

	salt := make([]byte, length)

	for k, v := range rnd {
		salt[k] = byte(alphabet[int(v)%len(alphabet)])
	}
	return salt, nil
}

func (u *User) SetPassword(password string) error {
	// Parts are:
	// algorithm$iterations$salt$hash
	var passwordEntry string
	if len(password) == 0 {
		passwordEntry = "pbkdf2_sha256$0$$!" // Cannot match "!"
	} else {
		salt, err := GenerateRandomString(12)
		if err != nil {
			return err
		}
		iterations := 24000

		hashed := pbkdf2.Key([]byte(password), salt, iterations, 32, sha256.New)
		hashEncoded := base64.StdEncoding.EncodeToString(hashed)
		passwordEntry = fmt.Sprintf("pbkdf2_sha256$%d$%s$%s", iterations, string(salt), hashEncoded)
	}

	u.password = passwordEntry

	return nil
}

func (u *User) ValidatePassword(password string) (bool, error) {
	// Non-existing users can never have valid passwords
	if u.ID == 0 {
		return false, nil
	}

	parts := strings.Split(u.password, "$")
	if len(parts) != 4 {
		return false, ErrInvalidHash
	}

	algorithm := parts[0]
	iterationsStr := parts[1]
	saltStr := parts[2]
	hashEncoded := parts[3]

	if algorithm != "pbkdf2_sha256" {
		return false, ErrUnknownPasswordAlgorithm
	}

	iterations, err := strconv.Atoi(iterationsStr)
	if err != nil {
		return false, zerr.Wrap(err).WithString("iterations", iterationsStr).WithAny("user", u)
	}

	if hashEncoded[0] == '!' {
		// Cannot match hash "!"
		return false, nil
	}

	suppliedHash := pbkdf2.Key([]byte(password), []byte(saltStr), iterations, 32, sha256.New)
	suppliedHashEncoded := base64.StdEncoding.EncodeToString(suppliedHash)

	if suppliedHashEncoded == hashEncoded {
		return true, nil
	}
	return false, nil
}

func UserSave(ctx context.Context, user User) (int, error) {
	tx := getContextTx(ctx)
	if user.ID > 0 {
		query := `UPDATE "user" SET 
name = $2,
password = $3,
email = $4
WHERE id = $1`
		_, err := tx.Exec(ctx, query, user.ID,
			user.Name,
			user.password,
			user.Email)
		if err != nil {
			return 0, zerr.Wrap(err).WithString("query", query).WithAny("user", user)
		}
		return user.ID, err
	}

	query := `INSERT INTO "user"
(username, email, name, password)
VALUES
($1, $2, $3, $4)
RETURNING id`

	err := tx.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.Name,
		user.password).Scan(&user.ID)
	if err != nil {
		return 0, zerr.Wrap(err).WithString("query", query).WithAny("user", user)
	}

	return user.ID, err
}

func UserGet(ctx context.Context, f UserFilter) (User, error) {
	query := `
SELECT id, username, email, name, password FROM "user"
`
	var filterstrings []string

	if f.ID > 0 {
		filterstrings = append(filterstrings, "id = :id")
	}

	if f.Username != "" {
		filterstrings = append(filterstrings, "LOWER(username) = LOWER(:username)")
	}

	if len(filterstrings) == 0 {
		return User{}, nil
	}

	var u User
	tx := getContextTx(ctx)
	rows, err := tx.NamedQuery(ctx, query, f)
	if err != nil {
		return u, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
	}
	defer rows.Close()

	for rows.Next() {
		tu := struct {
			User
			Password string
		}{}
		err = rows.StructScan(&tu)
		if err != nil {
			return u, zerr.Wrap(err).WithString("query", query).WithAny("filter", f)
		}
		tu.User.password = tu.Password
		u = tu.User
	}

	return u, nil
}
