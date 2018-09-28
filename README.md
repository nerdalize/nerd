# Nerd - Nerdalize Command Line Interface
Your personal `nerd` that takes care of running compute jobs on the [Nerdalize cloud](https://www.nerdalize.com/).

---

<img src="./nerd.svg">

[Nerdalize](https://www.nerdalize.com/) is building a different cloud. Instead of constructing huge datacenters, we're distributing our servers over homes. Homeowners use the residual heat for hot showers and to warm their house, and we don't need to build new infrastructure.

In order to make our cloud resources accessible and easy to use, we've developed a CLI that fits your workflow. Whether you’re a researcher, engineer or developer, it allows you to easily run your computations, simulations and analyses on our cloud infrastructure.

__Features__:
  - Moving __datasets__ from you workstation to the cloud and back is included right into the workflow
  - Nerd ensures efficient and quick datatransfers through a __deduplication__ algorithm
  - Send in __thousands of jobs__, Nerd makes sure your resources are used as efficiently as possible
  - Package your software using industry-standard __Docker__ containers
  - Follows basic CLI conventions to provide a __scriptable__ interface your daily dose of automation goodness

## Documentation
To start running your compute on the Nerdalize cloud you'll need to set up an account and download the Nerd CLI itself.

  - [Quickstarts](https://www.nerdalize.com/docs/) - To quickly get you up and running.
  - [Ready-to-use Software ](https://www.nerdalize.com/applications/) - We have application-specific guides for Python or FFmpeg for you to get started.
  - [CLI Reference](https://www.nerdalize.com/docs/reference/cli/) - For a reference of all available commands

---
## Building from Source
If you would like to contribute to the project it is possible to build the Nerd from source:

   1. The CLI is written in Go. Make sure you've installed the language SDK as documented [here](https://golang.org/dl/)
   2. Checkout the repository in your `GOPATH`:
      ```
      git clone git@github.com:nerdalize/nerd.git $GOPATH/src/github.com/nerdalize/nerd
      ```
   3. Go to the checked out repository and build the binary using the included bash script:
      ```
      cd $GOPATH/src/github.com/nerdalize/nerd
      ./make.sh build
      ```
   4. The Nerd CLI is now ready to be used in the `$GOPATH/bin` directory:
       ```
       $GOPATH/bin/nerd
       Usage: nerd [--version] [--help] <command> [<args>]

       Available commands are:
       ...
       ```

## Quickstart from Docker

Pull the docker image `nerdalize/nerd` then:

```
$ docker run -it --rm nerdalize/nerd -h
usage: nerd [--version] [--help] <command> [<args>]

Available commands are:
    dataset     upload and download datasets for tasks to use
    login       start a new authorized session
    project     set and list projects
    task        manage the lifecycle of compute tasks
    worker      control individual compute processes
    workload    control compute capacity for working on tasks
```
