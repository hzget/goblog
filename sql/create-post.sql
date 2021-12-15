DROP TABLE IF EXISTS post;
CREATE TABLE post(
  id        INT AUTO_INCREMENT NOT NULL,
  title     TINYTEXT NOT NULL,
  author    VARCHAR(10) NOT NULL, 
  ctime     DATETIME NOT NULL,
  mtime     DATETIME NOT NULL,
  body      LONGTEXT,
  PRIMARY KEY (`id`)
);
DROP TABLE IF EXISTS poststatistics;
CREATE TABLE poststatistics(
  postid    INT NOT NULL UNIQUE,
  star1    INT NOT NULL DEFAULT 0,
  star2    INT NOT NULL DEFAULT 0,
  star3    INT NOT NULL DEFAULT 0,
  star4    INT NOT NULL DEFAULT 0,
  star5    INT NOT NULL DEFAULT 0,
  PRIMARY KEY (`postid`)
);
