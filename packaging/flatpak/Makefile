.PHONY: vendor build install run

vendor:
	cd ../.. && GOWORK=off go mod vendor

build: vendor
	flatpak-builder --user --force-clean build-dir io.github.uzer007_admin.granio.flatpak_demo.yml

install: vendor
	flatpak-builder --user --install --force-clean build-dir io.github.uzer007_admin.granio.flatpak_demo.yml

run:
	flatpak run --user io.github.uzer007_admin.granio.flatpak_demo
