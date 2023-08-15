#!/bin/sh
BASE=$(pwd)
docker run -p 6001:8080 -e SWAGGER_JSON=/swagger/cluster_services.swagger.json -v ./api:/swagger swaggerapi/swagger-ui
