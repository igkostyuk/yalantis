

## Request Counter

`Request Counter` application - HTTP server that on each request responds with a counter of the total number of requests 
      
`Request Counter` store it's data in (https://redis.io/).


## HTTP interface
```

  GET http://localhost:3000/           return json:  { "count":3 }


  http://localhost:4000/debug/pprof    server runtime profiling data 


  http://localhost:4000/debug/vars     standardized interface to operation counters in server
  
```
    
## Installation
```
 # Building containers

 all: counter

 counter:
 	 docker build \
 		 -f docker/dockerfile.counter-api \
		 -t counter-api-amd64:1.0 \
		 --build-arg VCS_REF=`git rev-parse HEAD` \
		 --build-arg BUILD_DATE=`date -u +”%Y-%m-%dT%H:%M:%SZ”` \
		 .

 # Running from within docker compose

 run: up 

 up:
	 docker-compose -f docker/compose/compose.yaml  up --detach --remove-orphans

 down:
	 docker-compose -f docker/compose/compose.yaml down --remove-orphans

logs:
	 docker-compose -f docker/compose/compose.yaml logs -f


 # Modules support

 deps-reset:
	 git checkout -- go.mod
	 go mod tidy
	 go mod vendor

 tidy:
	 go mod tidy
	 go mod vendor

 deps-upgrade:
	 # go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	 go get -u -t -d -v ./...
	 go mod tidy
	 go mod vendor

 deps-cleancache:
	 go clean -modcache


 # Docker support

 FILES := $(shell docker ps -aq)

 down-local:
	 docker stop $(FILES)
	 docker rm $(FILES)

 clean:
	 docker system prune -f	

 logs-local:
	 docker logs -f $(FILES)
```

## Configuration

```

#The ip:port for the api endpoint.
api_host localhost:3000 

#The ip:port for the debug endpoint.
debug_host localhost:4000 

#The maximum duration for reading request.
read_timeout, 5*time.Second 

#The maximum duration before timing out writes of the response.
write_timeout, 5*time.Second 

#The maximum duration for stop server gracefully.
sutdown_timeout, 5*time.Second 
    
#The ip:port for the api endpoint.
db_host localhost:6379 

#The name for database
db_name ""

#The user name for database
db_user ""

#The password for database
db_password ""

```


## TODO
 - Add more unit and functional tests



