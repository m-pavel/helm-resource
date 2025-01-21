HELM_3_PLUGINS := $(shell helm env HELM_PLUGINS)


build:
		go build -o resource main.go

test:
		go test -v ./...
		
install: build
		mkdir -p $(HELM_3_PLUGINS)/helm-resource/bin
		cp resource $(HELM_3_PLUGINS)/helm-resource/bin
		cp plugin.yaml $(HELM_3_PLUGINS)/helm-resource/

