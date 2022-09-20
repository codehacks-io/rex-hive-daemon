<h1 align="center">Rex Hive Daemon</h1>

<p align="center">
  <img src="https://rex-public-assets.s3.amazonaws.com/rex-hive-logo.png" alt="logo" width="120px" height="120px"/>
  <br/>
  <i>
    Spawn a group of processes using a yaml file spec.
  </i>
  <br/>
</p>

<p align="center">
  <a href="https://gameship.io/rex-hive-api?src=github"><strong>gameship.io/rex-hive-api</strong></a>
  <br>
</p>

<p align="center">
  <a href="https://github.com/codehacks-io/rex-hive-api">
    <img src="https://img.shields.io/badge/version-0.0.0-brightgreen" alt="Version"/>
  </a>
</p>

<hr>

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

## Getting Started

Run mongo db with Docker:

```shell
docker run --name mongodb -d -v ./my_mongo_data/:/data/db -p 27017:27017 -e MONGO_INITDB_ROOT_USERNAME=rex-hive -e MONGO_INITDB_ROOT_PASSWORD=unsafe-password-NEVER-use-in-prod-1909 mongo:5.0.12
```

Run example spec:

```shell
go run . --file=./demo-specs/test-spec.yml
```

-----

## Test a go file in a remote machine

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

Check go installation

```shell
go version
```
