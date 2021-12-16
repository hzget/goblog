# goblog

A blog system used for sharing ideas and analyzing
these articles via AI algorithms.

## services

### work as a blog site

* read/write blogs
* signup/signin/logout
* vote with stars 1~5
* user ranks: bronze, silver, gold
* user admin

### work as code browsing platform

In the debug mode, programmers can browse underlying code on line.
It can help them to learn this system and make the debug life easier.

### work as AI analysis system

As a reader, the gold and silver user can
get AI analysis of article(s) on the blog.

As a programmer, the user can develop the AI functions.

this module is under developping now

## how to use

### Prerequisites

* mysql for storing blog posts
* redis for storing cache -- login sessions

You can change to others for corresponding service. Just only make very little code changes.

### run the code
Just run the command ***go run .*** :

```bash
$ go run .
/55e2e2fd-ae96-45b9-9249-6740416ebe18
PONG <nil>
Connected!

```

After that, visit the url via a web browser:

http://youripaddr:8080/55e2e2fd-ae96-45b9-9249-6740416ebe18/
