all: apple-bssid-enumerator

apple-bssid-enumerator: main.go wloc/*.go cmd/*.go proto/*.go cperm/*.go common/*.go constants/*.go 
	#protoc --go_out=. bssid.proto
	go mod tidy
	go build 

clean:
	rm apple-bssid-enumerator
