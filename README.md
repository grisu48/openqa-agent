# openqa-agent

Minimalistic cross-platform agent for openQA. This is an experiment to allow openQA to connect to non-Linux hosts 
via a new cross-platform backend and agent. This is the agent part.

The agent runs as a webserver that allows the backend to execute arbitrary commands and push/pull files to the host.

## How does this work?

The agent runs by default on port 8421. The agent requires a custom authenticaton token, provided via the `-t` flag. An authentication token is a secret, required to be allowed access to the agent. Each request needs to pass the token via the `Token` http header.

1. Compile the agent `go build ./...`
2. Run the agent with a custom authentication token `./agent -t TOKEN`
3. Perform commands against the REST API

## REST API

| Path | Method | Description |
|------|--------|-------------|
| `/health.json` | GET | Get agent health |
| `/exec` | POST | Run a command (see below) |
| `/file` | GET | Get a file from server (see below) |
| `/file` | POST | Push a file to server (see below) |

### Run a command

Use POST requests against the `/exec` endpoint to run custom commands. The body is expected to be a json object of the following kind:

```json
{
    "cmd":"executable",
    "shell":"optional_shell",
    "uid": 1000,
    "gid": 1000,
    "cwd":"/tmp",
    "timeout": 30,
}
```

The `cmd` argument is the only argument required. It defines the command to be executed.
The response is a json object with the following properties:

```json
{
  "cmd": "echo 'hello world'",
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
