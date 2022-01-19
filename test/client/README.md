# TEST CASE

## automate test case written by golang

just run the testcase under ***test/client*** directory with command ***go test***
```bash
# go test -v -bench=.
=== RUN   TestPressure
=== RUN   TestPressure/NParallelView
=== RUN   TestPressure/NParallelSave
=== RUN   TestPressure/NSequentialView
=== RUN   TestPressure/NSequentialSave
--- PASS: TestPressure (1.70s)
    --- PASS: TestPressure/NParallelView (0.55s)
    --- PASS: TestPressure/NParallelSave (0.55s)
    --- PASS: TestPressure/NSequentialView (0.30s)
    --- PASS: TestPressure/NSequentialSave (0.29s)
=== RUN   TestViewCases
--- PASS: TestViewCases (0.00s)
=== RUN   TestSaveCases
--- PASS: TestSaveCases (0.02s)
=== RUN   TestSaveAndViewCases
--- PASS: TestSaveAndViewCases (0.02s)
goos: linux
goarch: amd64
pkg: github.com/hzget/goblog/test/client
cpu: Intel(R) Core(TM) i5-7200U CPU @ 2.50GHz
BenchmarkViewN
BenchmarkViewN 	     709	   1647005 ns/op
PASS
ok  	github.com/hzget/goblog/test/client	4.059s
```

## manually test via curl test via curl

```bash
# create a post
$ curl http://127.0.0.1:8080/savejs -b cookies.txt -d '{"id":0, "title":"哦哈哟", "body":"骑上我心爱的小摩托，它永远不会堵车。"}'
# output:
{
	"success": true,
	"message": "save success",
	"id": 1
}

# view a post
$ curl -b cookies.txt 127.0.0.1:8080/viewjs -d '{"id":1}'
# output:
{
	"success": true,
	"message": "",
	"id": 1,
	"title": "哦哈哟",
	"author": "admin",
	"date": "2022-01-16T15:06:37Z",
	"modified": "2022-01-16T15:06:37Z",
	"body": "骑上我心爱的小摩托，它永远不会堵车。",
	"star": [
		0,
		0,
		0,
		0,
		0
	]
}
```
