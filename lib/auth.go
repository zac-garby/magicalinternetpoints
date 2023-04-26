package lib

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func (b *Backend) AuthLoginHandler(c *fiber.Ctx) error {
	email, password := c.FormValue("email"), c.FormValue("password")

	// validation
	if len(password) > 72 || len(password) < 5 {
		return fmt.Errorf("password must be between 5 and 72 characters")
	}

	// lookup email
	user, err := b.LookupEmail(email)
	if err != nil {
		return fmt.Errorf("could not find user")
	}

	// check correct password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(password)); err != nil {
		return fmt.Errorf("invalid password")
	}

	// successful; set session
	if err = b.BeginSession(user, c); err != nil {
		panic(err)
	}

	return c.Redirect("/")
}

func (b *Backend) AuthRegisterHandler(c *fiber.Ctx) error {
	var (
		username = c.FormValue("username")
		email    = c.FormValue("email")
		password = c.FormValue("password")
	)

	// validation
	if len(password) > 72 || len(password) < 5 {
		return fmt.Errorf("password must be between 5 and 72 characters")
	}

	if len(username) > 48 || len(password) < 3 {
		return fmt.Errorf("username must be between 3 and 48 characters")
	}

	if len(email) > 128 || len(email) < 4 {
		return fmt.Errorf("email must be between 4 and 128 characters")
	}

	// lookup email
	if _, err := b.LookupEmail(email); err == nil {
		return fmt.Errorf("user already exists with this email")
	}

	// lookup username
	if _, err := b.LookupUsername(email); err == nil {
		return fmt.Errorf("user already exists with this username")
	}

	// ok; create user
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		panic(err)
	}

	user := User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
	}

	b.CreateUser(user)

	if err = b.BeginSession(&user, c); err != nil {
		panic(err)
	}

	return c.Redirect("/")
}

func (b *Backend) AuthLogoutHandler(c *fiber.Ctx) error {
	if err := b.EndSession(c); err != nil {
		panic(err)
	}

	return c.Redirect("/")
}

func (b *Backend) LookupUserID(id int) (*User, error) {
	stmt, err := b.Storage.Conn().Prepare("SELECT id, email, username, password_hash FROM users WHERE id = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)
	user := new(User)

	if err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *Backend) LookupEmail(email string) (*User, error) {
	stmt, err := b.Storage.Conn().Prepare("SELECT id, email, username, password_hash FROM users WHERE email = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(email)
	user := new(User)

	if err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *Backend) LookupUsername(username string) (*User, error) {
	stmt, err := b.Storage.Conn().Prepare("SELECT id, email, username, password_hash FROM users WHERE username = ?")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(username)
	user := new(User)

	if err = row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash); err != nil {
		return nil, err
	}

	return user, nil
}

func (b *Backend) LookupAccounts(userID int) ([]*Account, error) {
	stmt, err := b.Storage.Conn().Prepare(`
		SELECT accounts.site_id, accounts.username, accounts.profile_url,
		       sites.title, sites.url, sites.score_description
		FROM accounts
		INNER JOIN sites ON accounts.site_id = sites.id
		WHERE accounts.user_id = ?
	`)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		account := new(Account)
		site := new(Site)
		if err = rows.Scan(
			&site.ID, &account.Username, &account.ProfileURL,
			&site.Title, &site.URL, &site.ScoreDescription,
		); err != nil {
			return nil, err
		}
		account.Site = *site
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

func (b *Backend) CreateUser(u User) error {
	// Get a reference to the underlying *sql.DB object
	db := b.Storage.Conn()

	// Define the SQL query
	query := "INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)"

	// Prepare the SQL statement
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	// Execute the SQL statement with the user data
	_, err = stmt.Exec(u.Email, u.Username, u.PasswordHash)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) BeginSession(user *User, c *fiber.Ctx) error {
	sess, err := b.Sessions.Get(c)
	if err != nil {
		return err
	}

	if !sess.Fresh() {
		if err = sess.Regenerate(); err != nil {
			return err
		}
	}

	sess.Set("id", user.ID)
	sess.SetExpiry(time.Hour * 24 * 7)

	return sess.Save()
}

func (b *Backend) EndSession(c *fiber.Ctx) error {
	sess, err := b.Sessions.Get(c)
	if err != nil {
		return err
	}

	return sess.Destroy()
}

func (b *Backend) CurrentUser(c *fiber.Ctx) (*User, error) {
	sess, err := b.Sessions.Get(c)
	if err != nil {
		return nil, err
	}

	v := sess.Get("id")
	if v == nil {
		return nil, fmt.Errorf("not signed in")
	}

	id, ok := v.(int)
	if !ok {
		panic("invalid session - user ID is not an int")
	}

	return b.LookupUserID(id)
}
