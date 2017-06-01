# Nerdalize Scientific Compute
Your personal nerd that takes care of running scientific compute on the [Nerdalize cloud](http://nerdalize.com/cloud/).

_NOTE: This project is currently experimental and not functional._

## Command Usage

```bash
# log into the scientific compute platform
$ nerd login
Successful login. You can now select a project using 'nerd project'

# list all projects
$ nerd project list
nerdalize-video
nerdalize-weather

# set a project to work with
$ nerd project set nerdalize-video

# upload a piece of data that will acts as input to the program
$ ls ~/Desktop/videos
video1.mov
video2.mov

$ nerd dataset upload ~/Desktop/videos
Uploading dataset with ID 'd-96fac377'
314.38 MiB / 314.38 MiB [=============================] 100.00%

# start a workload
# we start 2 workers that use the jrottenberg/ffmpeg container to work on the input dataset
$ nerd workload start jrottenberg/ffmpeg
    --instances 2
    --input-dataset d-96fac377
Started workload with ID 'w-96fac375'

# start two tasks for this workload
# this will start the jrottenberg/ffmpeg container twice with the given arguments
# input dataset d-96fac377 will be available in /input, data in /output will be uploaded when the task has successfully executed
$ nerd task start w-96fac375 -- -i /input/video1.mov -acodec copy -vcodec copy /output/video1.avi
$ nerd task start w-96fac375 -- -i /input/video2.mov -acodec copy -vcodec copy /output/video2.avi

# get status of tasks
$ nerd task list w-96fac375
+------------+------------+---------+
| WORKLOADID |   TASKID   | STATUS  |
+------------+------------+---------+
| w-96fac375 | t-14962176 | SUCCESS |
| w-96fac375 | t-89491732 | PENDING |
+------------+------------+---------+

# when all tasks are done we can download the output
$ nerd workload download w-96fac375 ~/Desktop/videos_out
$ tree ~/Desktop/videos_out
~/Desktop/videos_out
├── 7691e3df0c824efc1007082057b9c867_14962176
│   └── video1.avi
└── f45371445a7b95f7352bf841c30e4f58_89491732
    └── video2.avi

```

Please note that each command has a `--help` option that shows how to use the command.
Each command accepts at least the following options:
```
--config-file=  location of config file [$NERD_CONFIG_FILE]
--session-file= location of session file [$NERD_SESSION_FILE]
-v, --verbose=      show verbose output (default: false)
--json-format=  show output in json format (default: false)
```

## Power users

### Config

The `nerd` command uses a config file located at `~/.nerd/config.json` (location can be changed with the `--config-file` option) which can be used to customize nerd's behaviour.
The structure of the config and the defaults are show below:
```bash
{
        "auth": {
                "api_endpoint": "http://auth.nerdalize.com", # URL of authentication server
                "public_key": "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk\nkyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa\nOWbQHMK+vvUXieCJvCc9Vj084ABwLBgX\n-----END PUBLIC KEY-----", # Public key used to verify JWT signature
                "client_id": "GuoeRJLYOXzVa9ydPjKi83lCctWtXpNHuiy46Yux", #OAuth client ID
                "oauth_localserver": "localhost:9876", #address of local oauth server
                "oauth_success_url": "https://cloud.nerdalize.com" #redirect URL after successful login
        },
        "logging": { # write all command output to a log file
          "enabled": false,
          "file_location": "~/.nerd/log"
        },
        "nerd_api_endpoint": "https://batch.nerdalize.com" # URL of nerdalize API (NCE)
}
```

Session details such as OAuth session, JWT, and current project are stored in `~/.nerd/session.json` (can be changed using the `--session-file` option)
The structure of `session.json` is show below:
```bash
{
  "oauth": {
  	"access_token": "", #oauth access token
  	"refresh_token": "", #oauth refresh token
  	"expiration": "", #expiration date + time
  	"scope": "", #oauth scope
  	"token_type": "" #Bearer
  },
  "jwt": {
    "token": "", #Current JWT
    "refresh_token": "" #used when JWT is refreshable
  },
  "project": {
    "name": "", #NLZ project name
    "aws_region": "" #AWS Region
  }
}
```

## Docker

The nerd CLI can be dockerized. To build the docker container run:

```docker build -t my-nerd .```

You can now run the container like so:

```docker run my-nerd <command>```

If you want to use your local nerd config file (which contains your credentials), you can mount it:

```docker run -v ~/.nerd:/root/.nerd my-nerd <command>```

If you just want to set your credentials, you can also set it with an environment variable:

```docker run -e NERD_JWT=put.jwt.here my-nerd <command>```

## Nerdalize SDK

Code in this repository can also be used as a Software Development Kit (SDK) to communicate with Nerdalize services. The SDK consists of two packages:

### nerd/client

* `auth` is a client to the Nerdalize authentication backend. It can be used to fetch new JWTs.
* `batch` is a client to batch.nerdalize.com. It can be used to work with resources like `queues`, `workers`, and `datasets`.

### nerd/service

* `datatransfer` makes it possible to upload or download a dataset using one function call
* `working` works on workload tasks
