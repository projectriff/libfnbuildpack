# `riff-buildpack`
The Riff Buildpack is a Cloud Native Buildpack V3 that provides Riff Invokers to functions.

## Detection
Detection passes if a `$APPLICATION_ROOT/riff.toml` exists and the build plan already contains a `jvm-application` key.  If detection passes, the buildpack will contribute an `openjdk-jre` key with `launch` metadata to instruct the `openjdk-buildpack` to provide a JRE.  It will also add a `riff-invoker-java-` key and `handler` metadata extracted from the Riff metadata.

## Build
```toml
[riff-invoker-java]

  [riff-invoker-java.metadata]
  handler = "FQN of handler"
```

* Contributes Riff Java Invoker to a launch layer
* Contributes `web` process
* Contributes `riff` process

## License
This buildpack is released under version 2.0 of the [Apache License][a].

[a]: http://www.apache.org/licenses/LICENSE-2.0

