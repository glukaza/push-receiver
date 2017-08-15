.PHONY: all compile docker run

all:
	@echo '                                                                       '
	@echo -e '          \e[32m,""``.\e[0m     Hello my dear user                  '
	@echo -e '         \e[32m/ _  _ \\\e[0m    Makefile for your app              '
	@echo -e '         \e[32m|(\e[0m\e[31m@\e[0m\e[32m)(\e[0m\e[31m@\e[0m\e[32m)|\e[0m    This is Octopus'
	@echo -e '         \e[32m)  \e[31m~~\e[0m  \e[32m(\e[0m    He said: use make mother faka'
	@echo -e '        \e[32m/,`))((`.\\\e[0m'
	@echo -e '       \e[32m(( ((  )) ))\e[0m'
	@echo -e '        \e[32m`\ `)(` /`\e[0m '
	@echo -e '                                                                    '
	@echo '                                                                       '
	@echo 'DEFAULT:                                                               '
	@echo '   make compile                                                        '
	@echo '   make docker                                                         '
	@echo '   '
	@echo 'RUN:'
	@echo '   make run'

compile:
#	@git config --global url."git@github.com:".insteadOf "https://github.com/"
	go get -d -u github.com/glukaza/push-receiver
	go build

docker:
	@echo 'Build Docker'
	cp push-receivers docker/rootfs/opt/
ifdef VERSION
ifeq ($(VERSION), dev)
	cd docker && docker build -t docker-registry/ci/push-receivers:$(VERSION) .
	docker push docker-registry/ci/push-receivers:$(VERSION)
else
	cd docker && docker build -t docker-registry/ci/push-receivers:$(VERSION) .
	docker push docker-registry/ci/push-receivers:$(VERSION)
	docker tag docker-registry/ci/push-receivers:$(VERSION) docker-registry/ci/push-receivers:latest
	docker push docker-registry/ci/push-receivers:latest
endif
else
	cd docker && docker build -t hub.docker.com/gluka/push-receivers:latest .
endif

run:
	 docker run -it hub.docker.com/gluka/push-receivers:latest
