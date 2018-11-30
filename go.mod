module github.com/projectriff/riff-buildpack

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/buildpack/libbuildpack v1.6.0
	github.com/cloudfoundry/jvm-application-buildpack v1.0.0-M3
	github.com/cloudfoundry/libcfbuildpack v1.31.0
	github.com/cloudfoundry/nodejs-cnb v0.0.2
	github.com/cloudfoundry/npm-cnb v0.0.2
	github.com/cloudfoundry/openjdk-buildpack v1.0.0-M3
	github.com/onsi/gomega v1.4.3
	github.com/sclevine/spec v1.2.0
	golang.org/x/net v0.0.0-20181213202711-891ebc4b82d6 // indirect
	golang.org/x/sys v0.0.0-20181213200352-4d1cda033e06 // indirect
)

// TODO delete this line once we've updated libbuildpack
replace bou.ke/monkey v1.0.1 => github.com/bouk/monkey v1.0.0
