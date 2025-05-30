# openqa-agent

Minimalistic cross-platform agent for openQA. This is an experiment to allow openQA to connect to non-Linux hosts 
via a new cross-platform backend and agent. This is the agent part.

The agent runs as a webserver and/or serial port server that allows the backend to execute arbitrary commands and push/pull files to the host.

## How does this work?

The agent runs by default on port 8421. The agent requires a custom authenticaton token, provided via the `-t` flag. An authentication token is a secret, required to be allowed access to the agent. Each request needs to pass the token via the `Token` http header.

The agent can run also on a serial port. It accepts commands, which are either json-encoded `Job` objects or simple commands separated by a newline character.

`openqa-agent` can run on a serial terminal and/or as a webserver, but at least one of them must be active.

### Building

Compile the agent `go build ./...` or use the provided [Taskfile](https://taskfile.dev/)

```
$ task build
```

## REST API

Run the agent with a custom authentication token and and optional bind argument:

```
./openqa-agent -t TOKEN [-b 127.0.0.1:8421]
```

Then you can perform actions against the exposed REST API:

| Path | Method | Description |
|------|--------|-------------|
| `/health.json` | GET | Get agent health |
| `/exec` | POST | Run a command (see below) |
| `/file` | GET | Get a file from server (see below) |
| `/file` | POST | Push a file to server (see below) |

Most API endpoints require a `Token` item in the http header for authentication.

### Run a command

Use POST requests against the `/exec` endpoint to run custom commands (requires `Token` header for authentication).
The body is expected to be a json object of the following kind:

```json
{
    "cmd":"executable",
    "shell":"optional_shell",
    "uid": 1000,
    "gid": 1000,
    "cwd": "/tmp",
    "timeout": 30,
}
```

The `cmd` argument is the only argument required. It defines the command to be executed.
The response is a `Reply` json object, e.g.

```json
{
  "cmd": "echo 'hello world'",
  "shell": "bash",
  "runtime": 11,
  "ret": 0,
  "stdout": "hello world\n",
  "stderr": ""
}
```

### Push/Pull files

You can use the `/files` endpoint to push/pull files. The endpoint takes a `path` argument.
Use a GET request to pull a file and a POST request to push a file. The file is then in the http body.

e.g. to get the file `/home/geekotest/123.txt` you need to do a GET request against `/files?path=/home/geekotest/123.txt`.

## Discovery service

`openqa-agent` has an optional discovery function, which allows systems to probe for running openqa-agents.
If enabled, the agent listens on a UDP port for broadcast messages. It replies then with a predefined Discovery json object of the following kind:

```json
{"agent":"openqa-agent","status":"ok","token":"<user_defined_token>"}
```

It's recommended to use the same port for discovery as for the agent itself (one is udp, one is tcp). The `token` allows to distinguish between different agents on the same network.

## Serial terminal

`openqa-agent` can receive commands from a serial terminal, execute them and send a `Reply` json object back, e.g.

```json
{"cmd":"echo hello world","shell":"powershell","runtime":291,"ret":0,"stdout":"hello\r\nworld\r\n","stderr":""}
```

File push/pull is not supported via the serial terminal.

On Windows, the serial terminal is enabled by default. On Linux it is disabled by default to avoid conflicts with `getty`.

The serial terminal accepts either plain text commands or the same json objects as the REST API:

```json
{ "cmd":"executable","shell":"optional_shell","uid": 1000,"gid": 1000,"cwd": "/tmp","timeout": 30, }
```

The `cmd` parameter is required, all other parameters are optional.