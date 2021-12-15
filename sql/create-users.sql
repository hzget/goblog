DROP TABLE IF EXISTS users;
CREATE TABLE users (
  username  VARCHAR(10) NOT NULL,
  password  VARCHAR(1024) NOT NULL,
  rank  ENUM('bronze','silver','gold') NOT NULL,
  PRIMARY KEY (`username`)
);
