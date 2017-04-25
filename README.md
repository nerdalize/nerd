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

# list projects that you have access to
$ nerd project list
1. ItsGettingHotInHere
2. NerdsBeCrunching

# set the current project to be working on
$ nerd project set ItsGettingHotInHere

# create a queue that will be used as input to a workload
$ nerd queue create
Created queue with ID 'c-27bb24ae'

# add a task to the queue
$ nerd queue add c-27bb24ae nerdpowerrrr
Added task 'nerdpowerrrr' to queue 'c-27bb24ae'

$ create an empty dataset
$ nerd dataset create
Dataset created with ID 'd-96fac377'

# upload a piece of data that will acts as input to the program
$ nerd dataset append d-96fac377 ~/my-data
Uploading to dataset with ID 'd-96fac377'
314.38 MiB / 314.38 MiB [=============================] 100.00%

# start a new workload
$ nerd workload start nlz.io/my-org/my-program:v1.2 --queue c-27bb24ae --dataset d-96fac377 --no-of-workers 5
Started workload with ID 'w-321be912'
Output data will be available at dataset 'd-d1093e88f'

# get the workload status
$ nerd workload status w-321be912
Output dataset: d-d1093e88f
Queues: 
- c-27bb24ae: 1/1 processed

# download output of workload w-321be912
$ nerd dataset download d-d1093e88f
Downloading dataset with ID 'd-d1093e88f'
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

### Config

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

### Workers (local task execution)

When `nerd run` is used, the specified task is executed on one of Nerdalize's servers. It is also possible to test the execution of a task on a local machine first. To do this run the `nerd work` command. For this you will need to have Docker installed on your local machine.
```bash
$ nerd work
waiting for task ...
  received task 't-83dd21e' ... done!
  downloading input 'd-421a11f' ... done! (0KiB new, 120MiB total)
  running 't-83dd21e' (kubectl create t-83dd21e.yaml) ... done!
  uploading output ... done!
<- task 't-83dd21e' succeeded!
waiting for task ...
```

## Examples

* [Usage of docker container as base image](examples/docker-base-image)
