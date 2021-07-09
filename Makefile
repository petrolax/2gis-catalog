dbuser = "postgres"
pass = "12345"
dbname = "test"

all:
	go run *.go

params:
	go run *.go -dbuser $(dbuser) -pass $(pass) -dbname $(dbname)
