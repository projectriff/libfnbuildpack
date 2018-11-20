# `riff-buildpack`
The riff Buildpack is a Cloud Native Buildpack V3 that provides riff Invokers to functions.

It supports the following invokers:
- [java](https://github.com/projectriff/java-function-invoker)
- [node](https://github.com/projectriff/node-function-invoker)
- [command](https://github.com/projectriff/command-function-invoker)

This buildpack is designed to work in collaboration with other buildpacks, which are tailored to
support (and know how to build / run) languages supported by riff.

## In Plain English
In a nutshell, when combined with the other buildpacks present in the [riff builder](https://github.com/projectriff/riff-buildpack-group) what this means (and especially when dealing with the riff CLI which takes care of the creation of the `riff.toml` file for you):
* The presence of a `pom.xml` or `build.gradle` file will result in the compilation and execution of a java function, thanks to the [java invoker](https://github.com/projectriff/java-function-invoker)
  1. the `--handler` flag is optional in certain cases, as documented by the java invoker
* The presence of a `package.json` file and/or the fact that the `--artifact` flag points to a `.js` file will result in
  1. the `npm installation` of the function if applicable
  2. the execution as a node function thanks to the [node invoker](https://github.com/projectriff/node-function-invoker)
* The fact that the `--artifact` flag points to a file with the execute permission will result in the execution as a command function, thanks to the [command invoker](https://github.com/projectriff/command-function-invoker)
* Ambiguity in the detection process will result in a build failure
* The presence of the `--invoker` flag will entirely bypass the detection mechanism and force a given language/invoker 

## Detailed Buildpack Behavior

### Detection Phase
Detection passes if 
- a `$APPLICATION_ROOT/riff.toml` exists and 
- Either
    1. the build plan already contains a `jvm-application` key (typically because a JVM based application was detected by the [java buildpack](https://github.com/cloudfoundry/build-system-buildpack))
    2. the build plan already contains a `npm` key (typically because an NPM based application was detected by the [npm buildpack](https://github.com/cloudfoundry/npm-cnb))
        
        1. alternatively, if the file pointed to by the `artifact` value in `riff.toml` exists and has a `.js` extension
    3. as a fallback, the file pointed to by the `artifact` value in `riff.toml` exists and is executable
    
If detection passes in (i), the buildpack will contribute an `openjdk-jre` key with `launch` metadata to instruct 
the `openjdk-buildpack` to provide a JRE.  It will also add a `riff-invoker-java` key and `handler` 
metadata extracted from the riff metadata.

If detection passes in (ii), the buildpack will add a `riff-invoker-node` key and `fn` 
metadata extracted from the riff metadata.

If detection passes in (iii), the buildpack will add a `riff-invoker-command` key and `command` 
metadata extracted from the riff metadata.

If several languages are detected simultaneously, the detect phase errors out.
The `override` key in `riff.toml` can be used to bypass detection and force the use of a particular invoker.

### Build Phase

If a java build has been detected
* Contributes the riff Java Invoker to a launch layer, set as the main java entry point with `function.uri = <build-directory>?handler=<handler>` set as an environment variable.

If a node function has been detected
* Contributes the riff Node Invoker to a launch layer, set as the main `node` entry point with `FUNCTION_URI = <artifact>` set as an environment variable.
Note that `artifact` may actually be empty, in which case the invoker will `require()` the current directory (the function), which in turn expects that it contains a valid `package.json` file with its `main` entry point set. 

If a command function has been selected
* Contributes the riff Command Invoker to a launch layer, set as the main executable with `FUNCTION_URI = <artifact>` set as an environment variable.

In all cases, the function behavior is exposed _via_ standard buildpack [process types](https://github.com/buildpack/spec/blob/master/buildpack.md#launch):
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
This buildpack is released under version 2.0 of the [Apache License](http://www.apache.org/licenses/LICENSE-2.0).

