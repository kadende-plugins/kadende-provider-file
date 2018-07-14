build:
	rm -f plugin.so
	go build -buildmode=plugin -o plugin.so plugin.go