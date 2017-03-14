# Using nerd docker container as base image

When you want to run a container on the Nerdalize platform and want data management to be handled for you, the `nerdalize/nerd` container could be used as a convenient base container.
Running the `upload` and `download` command is up to the user. This example contains a minimal Dockerfile that shows how to use the `upload` and `download` commands.

Please note:
* The `/in` and `/out` folders in the container are the predefined locations to store input and output data.
* The `nerdalize/nerd` container expects a config file with a valid nerd token to be provided as a volume (e.g. `docker run -v ~/.nerd:/root/.nerd [YOUR-CONTAINER]`)
