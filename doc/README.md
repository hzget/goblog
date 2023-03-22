Readme
======

Protocls
--------

* http
* grpc

End Points
----------

/signup
/signin
/logout

/view/#id
/edit/#id

/viewjs
/savejs

SQL tables
----------

    mysql> show tables;
    +----------------+
    | Tables_in_blog |
    +----------------+
    | post           |
    | poststatistics |
    | users          |
    +----------------+
    3 rows in set (0.00 sec)

    mysql> 
    mysql> select * from post;
    +----+-------+--------+---------------------+---------------------+-------------+
    | id | title | author | ctime               | mtime               | body        |
    +----+-------+--------+---------------------+---------------------+-------------+
    |  1 | hello | admin  | 2023-03-21 16:58:03 | 2023-03-21 16:58:03 | how are you |
    +----+-------+--------+---------------------+---------------------+-------------+
    1 row in set (0.02 sec)

    mysql> 
    mysql> select * from poststatistics;
    +--------+-------+-------+-------+-------+-------+
    | postid | star1 | star2 | star3 | star4 | star5 |
    +--------+-------+-------+-------+-------+-------+
    |      1 |     0 |     0 |     1 |     0 |     1 |
    +--------+-------+-------+-------+-------+-------+
    1 row in set (0.00 sec)

    mysql> 
    mysql> select * from users;
    +----------+--------------------------------------------------------------+--------+
    | username | password                                                     | rank   |
    +----------+--------------------------------------------------------------+--------+
    | abc      | $2a$08$gCNG.F6ECACzFOiyc4ttoOP4/GVDma9zT5Cc/6n1jA6TxTDk3h6AS | silver |
    | admin    | $2a$08$YIo0VGfMcTNr0uCHA0pIH.JSgZ.hYgvKwicqKrwNa42FJffmeK8Dy | gold   |
    +----------+--------------------------------------------------------------+--------+
    2 rows in set (0.00 sec)

    mysql>         

create tables once connectint to the database

        CREATE TABLE IF NOT EXISTS post(
          id        INT AUTO_INCREMENT NOT NULL,
          title     TINYTEXT NOT NULL,
          author    VARCHAR(10) NOT NULL,
          ctime     DATETIME NOT NULL,
          mtime     DATETIME NOT NULL,
          body      LONGTEXT,
          PRIMARY KEY (id)
        );
        CREATE TABLE IF NOT EXISTS poststatistics(
          postid    INT NOT NULL UNIQUE,
          star1    INT NOT NULL DEFAULT 0,
          star2    INT NOT NULL DEFAULT 0,
          star3    INT NOT NULL DEFAULT 0,
          star4    INT NOT NULL DEFAULT 0,
          star5    INT NOT NULL DEFAULT 0,
          PRIMARY KEY (postid)
        );
        CREATE TABLE IF NOT EXISTS users (
          username  VARCHAR(10) NOT NULL,
          password  VARCHAR(1024) NOT NULL,
          `rank` + ENUM('bronze','silver','gold') NOT NULL,
          PRIMARY KEY (username)
        );
