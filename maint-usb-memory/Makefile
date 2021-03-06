
IMAGE_SIZE=20M

SETUP_DEV_LABEL=bme-setup
DEBUG_DEV_LABEL=bme-debug

# Setup用USBメモリに含めるファイルとディレクトリ
SETUP_FILES=led.py setup.conf setup.sh system-reset.sh
SETUP_DIRS=service/ config/ gpg/

# Debug用USBメモリに含めるファイルとディレクトリ
DEBUG_FILES=${SETUP_FILES}


.PHONEY: all clean

all: setup-usb.img setup-usb.img.gz debug-usb.img debug-usb.img.gz

clean:
	rm -f setup-usb.img setup-usb.img.gz debug-usb.img debug-usb.img.gz
	# remove temporary files
	sudo rm -rf .su-img.tmp .su-root.tmp .su-fsroot.tmp
	sudo rm -rf .du-img.tmp .du-root.tmp .du-fsroot.tmp

setup-usb.img.gz: setup-usb.img
	rm -f setup-usb.img.gz
	gzip --keep setup-usb.img

debug-usb.img.gz: debug-usb.img
	rm -f debug-usb.img.gz
	gzip --keep debug-usb.img

setup-usb.img: Makefile ${SETUP_FILES} ${SETUP_DIRS}
	sudo rm -rf .su-img.tmp .su-root.tmp .su-fsroot.tmp
	mkdir -p .su-root.tmp .su-fsroot.tmp

	cd service && make
	cp -a ${SETUP_FILES} ${SETUP_DIRS} .su-root.tmp/
	rm .su-root.tmp/service/Makefile
	rm .su-root.tmp/config/*.tmpl.conf
	[ -f .su-root.tmp/config/bme280d.conf ] || (echo "ERROR: bme280d.cnof file is not found." >&2; exit 1)
	[ -f .su-root.tmp/config/wpa_supplicant.conf ] || (echo "ERROR: wpa_supplicant.conf file is not found." >&2; exit 1)
	touch .su-root.tmp/host_passwd.list
	touch .su-root.tmp/host_secret.list

	cd .su-root.tmp && find -type f ! -name '*.list' -exec sha256sum '{}' '+' |sort >../.su-sha256sum
	cd .su-root.tmp && mv ../.su-sha256sum sha256sum
	cd .su-root.tmp && gpg --output sha256sum.sig --sign sha256sum

	fallocate -l ${IMAGE_SIZE} .su-img.tmp
	# create a partition
	echo "g\nn\n1\n\n\nw" |fdisk .su-img.tmp
	sudo losetup --show -P /dev/loop4 .su-img.tmp
	sudo mkfs.fat /dev/loop4p1
	sudo fatlabel /dev/loop4p1 ${SETUP_DEV_LABEL}
	sudo mount /dev/loop4p1 .su-fsroot.tmp

	sudo rsync -rlptv .su-root.tmp/ .su-fsroot.tmp/
	sudo chown -R root:root .su-fsroot.tmp

	df .su-fsroot.tmp
	sudo umount .su-fsroot.tmp
	sudo losetup -d /dev/loop4
	mv .su-img.tmp setup-usb.img
	sudo rm -rf .su-root.tmp .su-fsroot.tmp

debug-usb.img: Makefile ${DEBUG_FILES}
	sudo rm -rf .du-img.tmp .du-root.tmp .du-fsroot.tmp
	mkdir -p .du-root.tmp .du-fsroot.tmp

	cp -a ${DEBUG_FILES} .du-root.tmp/

	cd .du-root.tmp && find -type f -exec sha256sum '{}' '+' |sort >../.du-sha256sum
	cd .du-root.tmp && mv ../.du-sha256sum sha256sum
	cd .du-root.tmp && gpg --output sha256sum.sig --sign sha256sum

	fallocate -l ${IMAGE_SIZE} .du-img.tmp
	# create a pertition
	echo "g\nn\n1\n\n\nw" |fdisk .du-img.tmp
	sudo losetup --show -P /dev/loop5 .du-img.tmp
	sudo mkfs.fat /dev/loop5p1
	sudo fatlabel /dev/loop5p1 ${DEBUG_DEV_LABEL}
	sudo mount /dev/loop5p1 .du-fsroot.tmp

	sudo rsync -rlptv .du-root.tmp/ .du-fsroot.tmp/
	sudo chown -R root:root .du-fsroot.tmp

	df .du-fsroot.tmp
	sudo umount .du-fsroot.tmp
	sudo losetup -d /dev/loop5
	mv .du-img.tmp debug-usb.img
	sudo rm -rf .du-root.tmp .du-fsroot.tmp
