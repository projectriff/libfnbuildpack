# `riff-buildpack`

The riff Buildpack is a Cloud Native Buildpack V3 that provides support to riff Invoker Buildpacks.

This buildpack is designed to work in collaboration with other buildpacks, which are tailored to
support (and know how to build / run) languages supported by riff.

While this buildpack does not contribute to the build, it ensures that exactly one function invoker buildpack is detected for the function. It must be run after all other riff invoker buildpacks in the buildpack group.

## Detailed Buildpack Behavior

### Detection Phase

Detection errors if

- a `$APPLICATION_ROOT/riff.toml` exists and
- Either
  1. zero invoker-buildpack was detected
  2. more then one invoker-buildpack was detected

### Build Phase

Not applicable

## How to Build

You can build the buildpack by running

```bash
make
```

This will package (with pre-downloaded cache layers) the buildpack in the
`artifactory/io/projectriff/riff/io.projectriff.riff/latest` directory. That can be used as a `uri` in a `builder.toml`
file of a builder (see https://github.com/projectriff/riff-buildpack-group)

## License

This buildpack is released under version 2.0 of the [Apache License](https://www.apache.org/licenses/LICENSE-2.0).
