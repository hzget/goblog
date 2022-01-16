# TEST CASE

## automate test case written by golang

just run the testcase under ***test/client*** directory with command ***go test***
```bash
# go test -v
=== RUN   TestViewNtimes
--- PASS: TestViewNtimes (14.83s)
=== RUN   TestViewCases
--- PASS: TestViewCases (0.00s)
=== RUN   TestSaveAndViewNtimes
--- PASS: TestSaveAndViewNtimes (0.08s)
=== RUN   TestSaveAndViewCases
--- PASS: TestSaveAndViewCases (0.00s)
PASS
ok  	github.com/hzget/goblog/test/client	14.917s
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
