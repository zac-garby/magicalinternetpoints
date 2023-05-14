package backend

import "fmt"

func (b *Backend) GetUserID(username string) (int, error) {
	rows, err := b.Storage.Conn().Query(`
	SELECT id
	FROM users
	WHERE username = ?`, username)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}

		return id, nil
	}

	return -1, fmt.Errorf("user %s not found", username)
}
