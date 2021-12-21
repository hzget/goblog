SET NAMES 'utf8mb4';
SET CHARACTER SET utf8mb4;
ALTER DATABASE blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE TABLE IF NOT EXISTS post(
  id        INT AUTO_INCREMENT NOT NULL,
  title     TINYTEXT NOT NULL,
  author    VARCHAR(10) NOT NULL, 
  ctime     DATETIME NOT NULL,
  mtime     DATETIME NOT NULL,
  body      LONGTEXT,
  PRIMARY KEY (`id`)
);
CREATE TABLE IF NOT EXISTS poststatistics(
  postid    INT NOT NULL UNIQUE,
  star1    INT NOT NULL DEFAULT 0,
  star2    INT NOT NULL DEFAULT 0,
  star3    INT NOT NULL DEFAULT 0,
  star4    INT NOT NULL DEFAULT 0,
  star5    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`postid`)
);
CREATE TABLE IF NOT EXISTS users (
  username  VARCHAR(10) NOT NULL,
  password  VARCHAR(1024) NOT NULL,
  rank  ENUM('bronze','silver','gold') NOT NULL,
  PRIMARY KEY (`username`)
);
