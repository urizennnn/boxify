.PHONY: setup clean

ALPINE_TAR := alpine-minirootfs-3.19.0-x86_64.tar.gz
ROOTFS_DIR := /tmp/boxify-rootfs

setup:
	@echo "Creating rootfs directory at $(ROOTFS_DIR)..."
	mkdir -p $(ROOTFS_DIR)
	@echo "Extracting Alpine minirootfs..."
	tar -xzf $(ALPINE_TAR) -C $(ROOTFS_DIR)
	@echo "Alpine rootfs extracted to $(ROOTFS_DIR)"
	cp ./pkg/daemon/boxifyd.service /etc/systemd/system/
	go build -o boxify ./cmd/boxify/main.go
	go build -o boxifyd ./cmd/boxifyd/main.go
	cp ./boxifyd /usr/local/bin/boxifyd
	cp ./boxify /usr/local/bin/boxify
	chmod +x /usr/local/bin/boxifyd
	chmod +x /usr/local/bin/boxify



clean:
	@echo "Cleaning up rootfs directory..."
	rm -rf $(ROOTFS_DIR)
	@echo "Cleaned up $(ROOTFS_DIR)"

run:
	@echo "Building boxify"
	go build -o boxify ./cmd/boxify/main.go
	@echo "Boxify binary built"
	@echo "Starting application"
	sudo ./boxify run --memory 1m --cpu=1 
