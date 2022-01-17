# runner

```sh
# run the test for 10 seconds with 100 concurrent goroutines
go run main.go -c 100 -d 10  http://echo.jsontest.com/title/ipsum/content/blah

# 2522
```

```sh
# run the test for 10 seconds or until 1000 requests has been made with 100 concurrent goroutines
go run main.go -c 100 -r 1000 -d 10  http://echo.jsontest.com/title/ipsum/content/blah

# 1000
```
