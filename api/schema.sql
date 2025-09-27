-- Schema for the game database

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    SO INTEGER,
    pet_id INTEGER,
    FOREIGN KEY (SO) REFERENCES users(id),
    FOREIGN KEY (pet_id) REFERENCES pets(id)
);

CREATE TABLE IF NOT EXISTS pets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    main_owner INTEGER NOT NULL,
    owner2 INTEGER NOT NULL,
    money INTEGER NOT NULL CHECK(money >= 0),
    health INTEGER CHECK(health BETWEEN 1 AND 100) DEFAULT 100,
    hunger INTEGER CHECK(hunger BETWEEN 1 AND 100) DEFAULT 100,
    happiness INTEGER CHECK(happiness BETWEEN 1 AND 100) DEFAULT 100,
    FOREIGN KEY (main_owner) REFERENCES users(id),
    FOREIGN KEY (owner2) REFERENCES users(id)
);