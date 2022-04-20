a: gen run

gen:
	openssl req  -new  -newkey rsa:2048  -nodes  -keyout localhost.key  -out localhost.csr -subj "/CN=localhost"
	openssl  x509  -req  -days 365  -in localhost.csr  -signkey localhost.key  -out localhost.crt

run:
	go run main.go
