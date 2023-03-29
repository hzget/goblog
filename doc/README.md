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
/vote
/analyze

/viewjs
/savejs

/superadmin
/saveranks

Validation
----------

makes use of http cookies to carry user info
and use redis to store the logged in session

Performance
-----------

       goos: linux
       goarch: amd64
       pkg: github.com/hzget/goblog/blog
       cpu: Intel(R) Core(TM) i5-7200U CPU @ 2.50GHz
       BenchmarkSigninWrapper
       BenchmarkSigninWrapper        39      28470708 ns/op       22471 B/op        202 allocs/op
       BenchmarkViewHandler
       BenchmarkViewHandler          46      28989517 ns/op       67169 B/op        676 allocs/op
       BenchmarkViewjs
       BenchmarkViewjs              100      12665127 ns/op       13074 B/op        128 allocs/op
       PASS
       ok      github.com/hzget/goblog/blog    6.501s

cache
-----

Make use of redis cache for mysql operation.

* reduce time consuming
* support more parallel read operation (more than mysql limit)

       phz@2004:~/proj/github.com/hzget/goblog/blog$ benchstat nocache.txt withcache.txt
       nocache.txt:8: missing iteration count
       withcache.txt:8: missing iteration count
       goos: linux
       goarch: amd64
       pkg: github.com/hzget/goblog/blog
       cpu: Intel(R) Core(TM) i5-7200U CPU @ 2.50GHz
              │ nocache.txt  │            withcache.txt            │
              │    sec/op    │   sec/op     vs base                │
       Viewjs   13.079m ± 5%   6.288m ± 6%  -51.92% (p=0.000 n=20)

              │ nocache.txt  │            withcache.txt            │
              │     B/op     │     B/op      vs base               │
       Viewjs   13.07Ki ± 0%   12.49Ki ± 0%  -4.44% (p=0.000 n=20)

              │ nocache.txt │           withcache.txt            │
              │  allocs/op  │ allocs/op   vs base                │
       Viewjs   130.00 ± 0%   92.00 ± 0%  -29.23% (p=0.000 n=20)

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
