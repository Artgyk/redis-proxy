test:
	docker-compose down
	docker-compose build
	docker-compose run --rm test
	docker-compose down
