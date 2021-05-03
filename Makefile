
.PHONY: deb

GITVERSION := $(shell git describe --tags --long 2>/dev/null)
ifeq ($(GITVERSION),)
    GITVERSION := 0-testing
endif

build:
	cd cmd/docsis-pnm && go build

deb: build
	mkdir -p debian/docsis-pnm/usr/bin || true
	cp cmd/docsis-pnm/docsis-pnm debian/docsis-pnm/usr/bin
	sed -e "s/^Version: .*/Version: $(GITVERSION)/g" <debian/docsis-pnm/DEBIAN/control.templ >debian/docsis-pnm/DEBIAN/control
	cd debian && dpkg -b docsis-pnm
	cd debian && mv docsis-pnm.deb docsis-pnm-$(GITVERSION)_$(shell dpkg --print-architecture).deb

clean:
	-rm -r debian/docsis-pnm/usr
	-rm cmd/docsis-pnm/docsis-pnm
	-rm debian/*.deb
