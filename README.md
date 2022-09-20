# Rex Hive Daemon

Spawn a group of processes using a yaml file spec.

## Glossary

- `HiveSpec`: Formal definition of how one or multiple processes will run in a machine. It's defined using `yaml`'.
- `HiveRun`: Group of processes running in a machine product of executing a `HiveSpec` by the `Rex Hive Daemon`. It's
assigned an ID once registered in DB.
- `Rex Hive Daemon`: Executable binary produced by this repo which is capable of reading a `HiveSpec` file and produce a
`HiveRun` from it.

## Motivation

In multiplayer games, we end up with 2 executables: The one that's executed in the player's console, and the one that's
executed in the server to enable multiplayer capabilities.

To optimize the usage of resources, we can instantiate the same game server multiple times within the same VM.

This project aims to allow to instantiate one or more processes using a yaml file definition, and monitor the stdout and
stderr channels of each process to ease debugging of multiplayer game deployments.

## Local dev

Run example spec:

```shell
go run . --file=./demo-specs/test-spec.yml
```

-----

You can easily run mongo db locally using Docker:

```shell
docker run --name mongodb -d -v ./my_mongo_data/:/data/db -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=rex-hive -e MONGO_INITDB_ROOT_PASSWORD=unsafe-password-NEVER-use-in-prod-1909 mongo:5.0.12
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

