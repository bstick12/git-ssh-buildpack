# GIT SSH key buildpack for Github

This is a [Cloud Native Buildpack V3](https://buildpacks.io/) that enables adding an SSH key to the build process and forces git to use SSH rather than HTTPS for communication with Github.

This buildpack is designed to work in collaboration with other buildpacks.

## Usage

Export your SSH key to an environment variable named `GIT_SSH_KEY`

```
export GIT_SSH_KEY=`cat /path/to/ssh_key`
```

Add the buildpack e.g via the pack cli

```
pack build <image-name> --builder cloudfoundry/cnb:cflinuxfs3 --buildpack /path/to/git-ssh-buildpack -e GIT_SSH_KEY ....
```

## Development

`scripts/unit.sh` - Runs unit tests for the buildpack
`scripts/build.sh` - Builds the buildpack
`scripts/package.sh` - Package the buildpack
