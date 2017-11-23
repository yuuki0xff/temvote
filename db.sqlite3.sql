CREATE TABLE session (
  session_id    INTEGER PRIMARY KEY AUTOINCREMENT,
  secret_sha256 CHAR(64) NOT NULL, -- COMMENT '16進数表記',
  expire        DATETIME NOT NULL
);

CREATE TABLE room (
  room_id       INTEGER PRIMARY KEY AUTOINCREMENT,
  name          TEXT NOT NULL, -- 'ex: 研A402',
  building_name TEXT NOT NULL, -- 'ex: 研究棟A',
  floor         INT  NOT NULL  -- '地下階はマイナスの値、地上階はプラスの値。0は存在しない'
);

CREATE TABLE thing (
  thing_id     INTEGER PRIMARY KEY AUTOINCREMENT,
  room_id      INTEGER             NOT NULL,
  thing_name   CHAR(32)            NOT NULL,
  update_cycle INTEGER DEFAULT 60  NOT NULL, -- '単位: 秒',

  UNIQUE (room_id, thing_name),
  FOREIGN KEY (room_id) REFERENCES room (room_id)
    ON DELETE CASCADE
);

CREATE TABLE vote (
  vote_id    INTEGER  PRIMARY KEY AUTOINCREMENT,
  session_id INTEGER  NOT NULL,
  room_id    INTEGER  NOT NULL,
  choice     CHAR(10) NOT NULL, -- 'hot, comfort, coldのいずれか',
  timestamp  DATETIME NOT NULL, -- '投票時刻',

  FOREIGN KEY (session_id) REFERENCES session (session_id)
    ON DELETE CASCADE,
  FOREIGN KEY (room_id) REFERENCES room (room_id)
    ON DELETE CASCADE
);

