.PHONY: all gluster-csi-driver clean

all: gluster-csi-driver

test:
	#go test github.com/gluster/gluster-csi-drivers/file/driver/... -cover
	#go vet github.com/gluster/gluster-csi-drivers/file/driver/...

gluster-csi-driver:
	if [ ! -d ./vendor ]; then dep ensure -vendor-only; fi
	go build -o build/gluster-csi-driver  main.go
clean:
	go clean -r -x
	-rm -rf build
