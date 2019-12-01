
all: mysqldb migrate
mysqldb:
	docker-compose up -d mysqldb

migrate:
	docker-compose run --rm migratedb

ps stop rm top:
	docker-compose $@
