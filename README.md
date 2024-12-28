# Manty-DNS
* simple dns server that is developed with Go Lang.

## How to use.
### start dns server
```shell
go run cmd/main.go 
DNS server is running on port 10054
```
### resolve dns
* You can use dig or nslookup to resolve .
```shell
 dig @127.0.0.1 -p 10054 www.naver.com
```
