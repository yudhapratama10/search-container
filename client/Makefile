build:
	go build

run:
	echo " >> run docker compose"
	echo " >> build client"
	make build
	echo " >> client built"
	echo " >> executing client"
	./client
	echo " >> client is running"

docker-build:
	docker-compose up --build
	echo " >> docker compose executed"

docker-down:
	docker-compose down
