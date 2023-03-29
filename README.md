goblog
======

A blog system used for recording ideas and analyzing
these articles via AI algorithms.

Architecture
------------

                            +--------+
    WebUI        ---->      |        |  ---->  mysql
                            |        |
    pprof        ---->      | goblog |  ---->  redis
                            |        |
    Test Suite   ---->      |        |  ---->  data center (AI analysis)
                            +--------+

WebUI
-----

view a post | edit a post | analysis | analyze one post | analysis result|
:----------:|:-----------:|:-------:|:------:|:------:
![view][view]|![edit][edit]|![analysis][analysis]|![rawpost][rawpost]|![result][result]

[Tech Detail](./doc)
-------------

It contains endpoints, performance, cache, sql tables, etc.

TO DO LIST
----------

* [x] machine learning code
* [x] basic test case via go
* [x] record leveled logs to a file
* [x] support middleware
* [x] support cache for mysql
* [ ] rate limit of the visit
* [x] access limit of database
* [ ] pprof investigation
* [ ] fix bottleneck of the database create/update operation
* [ ] garbage collection performance
* [ ] other performance issue

Prerequisites
-------------

* mysql: 8.0.27
* redis: 6.2.6
* hzget/analysis: 1.0

You can change to others for corresponding service. Just only make very little code changes.

Configuration
-------------

[config/config.json](./blog/config/config.json)

Run the code from within the host
---------------------------------

If prerequisites services are available,
modify the config file and run the command ***go run .*** :

       $ go build
       $ ./goblog
       PONG <nil>
       Connected!

After that, visit the url via a web browser: `http://localhost:8080/`

Run the code from within a container
------------------------------------

The user can run containers of goblog, mysql, redis and analysis.

* run all containers with docker-compose.yaml: `docker-compose up -d goblog`
* get goblog logs `docker-compose logs -f goblog`

Testing
-------

Test cases are in these `*_test.go` files inside [blog](./blog) dir.
The user can run with ***go test*** commands to test specific funcs
or run a benchmark test.

There're also test case by which the client start a real http connection.

* automate [test case](./test/client)
* manually test via ***curl*** command line

[view]: ./pic/view.png
[edit]: ./pic/edit.png
[analysis]: ./pic/analysis.png
[rawpost]: https://github.com/hzget/hzget.github.io/blob/feature/neural_networks/pics/analysis_raw.png
[result]: https://github.com/hzget/hzget.github.io/blob/feature/neural_networks/pics/analysis_result.png
