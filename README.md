# Swarm Chan

Spawn a swarm of processes using a yaml spec.

Useful to initialize and monitor server-side game instances.

## Local dev

You can easily run mongo db locally using Docker:

```shell
docker run --name mongodb --rm -d -v ./my_mongo_data/:/data/db -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=SwarmChan -e MONGO_INITDB_ROOT_PASSWORD=superSwarmChan-hunter2 mongo:5.0.12
```
