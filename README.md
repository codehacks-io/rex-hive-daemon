# Swarm Chan

Spawn a swarm of processes using a yaml spec.

Useful to initialize and monitor server-side game instances.

## Local dev

Run example swarm:

```shell
go run . --file=./demo-specs/test-spec.yml
```

-----

You can easily run mongo db locally using Docker:

```shell
docker run --name mongodb -d -v ./my_mongo_data/:/data/db -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=SwarmChan -e MONGO_INITDB_ROOT_PASSWORD=superSwarmChan-hunter2 mongo:5.0.12
```

# Test a go file in a remote machine

```shell
ssh -l ubuntu 3.91.99.67 -i .\my-meow-key
```

Download go binaries

```shell
wget -O ~/go-stuff.tar.gz https://go.dev/dl/go1.19.1.linux-amd64.tar.gz
```

Delete all go prev installations and unzip new go binaries

```shell
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf ~/go-stuff.tar.gz
```

Add /usr/local/go/bin to the PATH environment variable.

```shell
export PATH=$PATH:/usr/local/go/bin
```

Test go installation

```shell
go version
```

