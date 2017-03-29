# Nerdalize Scientific Compute
Your personal nerd that takes care of running scientific compute on the [Nerdalize cloud](http://nerdalize.com/cloud/).

_NOTE: This project is currently experimental and not functional._

## Command Usage

```bash
# log into the scientific compute platform
$ nerd login
Please enter your Nerdalize username and password.
Username: my-user@my-organization.com
Password: ******

# upload a piece of data that will acts as input to the program
$ nerd upload ./my-project/my-task-input
Uploading dataset with ID 'd-96fac377'
314.38 MiB / 314.38 MiB [=============================] 100.00%

# create a new task that takes the uploaded dataset as input
$ nerd run nlz.io/my-org/my-program:v1.2 d-96fac377
Created task with ID t-afb5cb16

# read task output to get feedback
$ nerd logs t-afb5cb16
20170122.1111 [INFO] Started program
20170122.2111 [INFO] Doing awesome science!

# start working the started item(s)
$ nerd work
waiting for task ...
  received task 't-83dd21e' ... done!
  downloading input 'd-421a11f' ... done! (0KiB new, 120MiB total)
  running 't-83dd21e' (kubectl create t-83dd21e.yaml) ... done!
  uploading output ... done!
<- task 't-83dd21e' succeeded!
waiting for task ...

# get each task's status
$ nerd status
|   TASKID   | OUTPUT DATASET |    CREATED     |
|------------|----------------|----------------|
| t-afb5cb16 | d-615f2d56     | 22 minutes ago |

# download results of running the task
$ nerd download t-615f2d56 ./my-project/my-task-output
Downloading dataset with ID 'd-615f2d56'
12.31 MiB / 12.31 MiB [=============================] 100.00%
```

Please note that each command has a `--help` option that shows how to use the command.
Each command accepts at least the following options:
```
      --config=      location of config file [$CONFIG]
  -v, --verbose      show verbose output (default: false)
      --json-format  show output in json format (default: false)
```

## Power users
The `nerd` command uses a config file located at `~/.nerd/config.json` (location can be changed with the `--config` option) which can be used to customize nerd's behaviour.
The structure of the config and the defaults are show below:
```bash
{
        "auth": {
                "api_endpoint": "http://auth.nerdalize.com", # URL of authentication server
                "public_key": "-----BEGIN PUBLIC KEY-----\nMHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk\nkyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa\nOWbQHMK+vvUXieCJvCc9Vj084ABwLBgX\n-----END PUBLIC KEY-----" # Public key used to verify JWT signature
        },
        "enable_logging": false, # When set to true, all output will be logged to ~/.nerd/log
        "current_project": "", # Current project
        "nerd_token": "", # Nerdalize JWT (can be set manually or it will be set by `nerd login`)
        "nerd_api_endpoint": "https://batch.nerdalize.com" # URL of nerdalize API (NCE)
}
```

## Examples

* [Usage of docker container as base image](examples/docker-base-image)
