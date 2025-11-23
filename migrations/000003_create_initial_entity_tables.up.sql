CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tournaments (
    id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('draft', 'started', 'completed')),
    tournament_type TEXT NOT NULL CHECK(tournament_type IN ('single', 'double')),
    score_requirement INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE entries (
    id TEXT PRIMARY KEY,
    tournament_id TEXT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    seed INTEGER NOT NULL,
    embed_link TEXT
);

CREATE TABLE matches (
    id TEXT PRIMARY KEY,
    tournament_id TEXT NOT NULL REFERENCES tournaments(id) ON DELETE CASCADE,

    bracket_side TEXT NOT NULL CHECK(bracket_side IN ('winners', 'losers', 'finals')),
    round_number INTEGER NOT NULL,
    match_order INTEGER NOT NULL,

    entry_1_id TEXT REFERENCES entries(id),
    entry_2_id TEXT REFERENCES entries(id),

    score_1 INTEGER NOT NULL DEFAULT 0,
    score_2 INTEGER NOT NULL DEFAULT 0,
    
    status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'finished')),
    winner_next_match_id TEXT REFERENCES matches(id),
    winner_next_slot INTEGER,

    loser_next_match_id TEXT REFERENCES matches(id),
    loser_next_slot INTEGER,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_entries_tournament ON entries(tournament_id);
CREATE INDEX idx_matches_tournament ON matches(tournament_id);

INSERT INTO users (id, email, username)
VALUES (
    '00000000-0000-0000-0000-000000000001', 
    'mock_user@fakeemail.lol', 
    'Mockinator'
);