# Nerdalize Scientific Compute
Your personal nerd that takes care of running scientific compute on the [Nerdalize cloud](http://nerdalize.com/cloud/).

_NOTE: This project is currently experimental and not functional._

## Command Usage

```bash
# log into the scientific compute platform
$ nerd login                              
username: my-user@my-organization.com
password: ******

# upload a piece of data that will acts as input to the program
$ nerd upload ./my-project/my-task-input
uploading input... done! (512KiB new, 120MiB total)
upload registered as dataset 'd-421a11f'

# create a new task that takes the uploaded dataset as input
$ nerd run nlz.io/my-org/my-program:v1.2 d-421a11f
creating task ... done!
submitted run as task 't-83dd21e'

# read task output to get feedback
$ nerd logs t-83dd21e
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

# download results of running the task
$ nerd download t-83dd21e ./my-project/my-task-output
downloading output... done! (100MiB)
```
