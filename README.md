# `riff-buildpack`
The riff Buildpack is a Cloud Native Buildpack V3 that provides riff Invokers to functions.

It supports the following invokers:
- [java](https://github.com/projectriff/java-function-invoker)
- [command](https://github.com/projectriff/command-function-invoker)

This buildpack is designed to work in collaboration with other buildpacks, which are tailored to
support (and know how to build / run) languages supported by riff.


## Detection Phase
Detection passes if 
- a `$APPLICATION_ROOT/riff.toml` exists and 
- Either
  1. the build plan already contains a `jvm-application` key
  2. as a fallback, the file pointed by the `artifact` value in `riff.toml` exists and is executable
    
If detection passes in (i), the buildpack will contribute an `openjdk-jre` key with `launch` metadata to instruct 
the `openjdk-buildpack` to provide a JRE.  It will also add a `riff-invoker-java` key and `handler` 
metadata extracted from the riff metadata.

If detection passes in (ii), the buildpack will add a `riff-invoker-command` key and `command` 
metadata extracted from the riff metadata.

## Build Phase

If a java build has been detected
* Contributes riff Java Invoker to a launch layer, set as the main java entry point with `function.uri = <build>?handler=<handler` set

If a command function has been selected
* Contributes the riff Command Invoker to a launch layer, set as the main executable with `FUNCTION_URI = <artifact>`


* Contributes `web` process
* Contributes `function` process

## How to Build

You can build the buildpack by running 
```bash
make
```

This will package (with pre-downloaded cache layers) the buildpack in the 
`scratch/io/projectriff/riff/io.projectriff.riff/latest` directory. That can be used as a `uri` in a `builder.toml`
file of a builder (see https://github.com/projectriff/riff-buildpack-group)


## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

