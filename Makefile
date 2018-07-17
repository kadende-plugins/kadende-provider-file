debug:
	rm -f /Users/francismwangi/go/src/github.com/kadende/cluster-controller/plugins/provider/file_0.0.1.so
	cp plugin.so /Users/francismwangi/go/src/github.com/kadende/cluster-controller/plugins/provider/file_0.0.1.so

build:
	rm -f plugin.so
	go build -buildmode=plugin -o plugin.so plugin.go