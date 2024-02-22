CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash BLOB NOT NULL);
CREATE TABLE accounts (
    user_id INTEGER,
    site_id INTEGER,
    username TEXT,
    profile_url TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (site_id) REFERENCES sites(id),
    PRIMARY KEY (user_id, site_id)
);
CREATE TABLE point_sources (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    site_id INTEGER, low_upper INTEGER, medium_upper INTEGER, low_rate REAL, medium_rate REAL, high_rate REAL,
    FOREIGN KEY (site_id) REFERENCES sites(id)
);
CREATE TABLE raw_points (
    user_id INTEGER,
    point_source_id INTEGER,
    point_total INTEGER NOT NULL DEFAULT 0,
    last_updated_date INTEGER NOT NULL, real_points INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, point_source_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (point_source_id) REFERENCES point_sources(id)
);
CREATE TABLE IF NOT EXISTS "sites" (
	"id"	INTEGER,
	"title"	TEXT,
	"url"	TEXT,
	"score_description"	TEXT NOT NULL DEFAULT '',
	PRIMARY KEY("id")
);
