all: 01 02

01: 01.go
	go build 01.go

02: 02.go
	go build 02.go

clean:
	rm -f 01 02
