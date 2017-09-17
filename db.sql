CREATE TABLE session (
  session_id    BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  secret_sha256 CHAR(64) NOT NULL COMMENT '16進数表記',
  expire        DATETIME NOT NULL
);

CREATE TABLE room (
  room_id       BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  name          TEXT NOT NULL COMMENT 'ex: 研A402',
  building_name TEXT NOT NULL COMMENT 'ex: 研究棟A',
  floor         INT  NOT NULL COMMENT '地下階はマイナスの値、地上階はプラスの値。0は存在しない'
) CHARSET = 'utf8';

CREATE TABLE thing (
  thing_id     BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  room_id      BIGINT UNSIGNED          NOT NULL,
  thing_name   CHAR(32)                 NOT NULL,
  update_cycle INT UNSIGNED DEFAULT 60  NOT NULL COMMENT '単位: 秒',

  UNIQUE (room_id, thing_name),
  FOREIGN KEY (room_id) REFERENCES room (room_id)
    ON DELETE CASCADE
);

CREATE TABLE vote (
  vote_id    BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  session_id BIGINT UNSIGNED NOT NULL,
  room_id    BIGINT UNSIGNED NOT NULL,
  choice     CHAR(10)        NOT NULL COMMENT 'hot, comfort, coldのいずれか',
  timestamp  DATETIME        NOT NULL COMMENT '投票時刻',

  FOREIGN KEY (session_id) REFERENCES session (session_id)
    ON DELETE CASCADE,
  FOREIGN KEY (room_id) REFERENCES room (room_id)
    ON DELETE CASCADE
);

