install-air: 
	curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s
	./bin/air -v

prepare:
	echo "Preparing project"
	make install-air

air: 
	./bin/air -c .air.toml

hello:
	echo "hello"

ls-port:
	lsof -n -i TCP:8000

kill-port:
	kill -9 $(lsof -ti:8000)

all: hello air