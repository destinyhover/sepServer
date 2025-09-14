DROP TABLE IF EXISTS users CASCADE;
CREATE TABLE IF NOT EXISTS users(
    UserID BIGINT GENERATED ALWAyuS AS IDENTITY PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    lastlogin BIGINT,
    admin SMALLINT NOT NULL DEFAULT 0, 
    active SMALLINT NOT NULL DEFAULT 0
);

INSERT INTO users (username, password, lastlogin, admin, active) VALUES ('admin', 'admin', 1757508731, 1,1);