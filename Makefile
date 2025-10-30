.PHONY: setup clean

ALPINE_TAR := alpine-minirootfs-3.19.0-x86_64.tar.gz
ROOTFS_DIR := /tmp/alpine-rootfs

setup:
	@echo "Creating rootfs directory at $(ROOTFS_DIR)..."
	mkdir -p $(ROOTFS_DIR)
	@echo "Extracting Alpine minirootfs..."
	tar -xzf $(ALPINE_TAR) -C $(ROOTFS_DIR)
	@echo "Alpine rootfs extracted to $(ROOTFS_DIR)"

clean:
	@echo "Cleaning up rootfs directory..."
	rm -rf $(ROOTFS_DIR)
	@echo "Cleaned up $(ROOTFS_DIR)"
