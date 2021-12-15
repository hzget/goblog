package main

/*
 * A blog system
 *
 *  Functions:
 *
 *          [*] writing blogs
 *          [*] singup/singin/logout
 *          [*] voting from readers
 *          [*] manage users: assign ranks to users
 *          [*] add random prefix for the url path
 *          [*] data analysis
 *          [*] view underying code -- make an easy life for programmers
 *
 */

import (
	"github.com/hzget/goblog/blog"
)

func main() {
	blog.Run(":8080")
}
