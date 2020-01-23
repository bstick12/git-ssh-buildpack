# GIT SSH key buildpack for Github/Bitbucket

This is a [Cloud Native Buildpack V3](https://buildpacks.io/) that enables adding an SSH key to the build process and forces git to use SSH rather than HTTPS for communication with [Github](https://github.com) or [Bitbucket](https://bitbucket.org). It requires [Go v1.13](https://golang.org).

This buildpack is designed to work in collaboration with other buildpacks. It is compatible to API v0.2 and it supports the stacks "org.cloudfoundry.stacks.cflinuxfs3" and "io.buildpacks.stacks.bionic". It has been tested with the [pack CLI](https://github.com/buildpack/pack) v0.5.0.

[![CircleCI](https://circleci.com/gh/bstick12/git-ssh-buildpack.svg?style=svg)](https://circleci.com/gh/bstick12/git-ssh-buildpack)
[![Download](https://api.bintray.com/packages/bstick12/buildpacks/git-ssh-buildpack/images/download.svg?version=0.1.0) ](https://bintray.com/bstick12/buildpacks/git-ssh-buildpack/0.1.0/link)
[![codecov](https://codecov.io/gh/bstick12/git-ssh-buildpack/branch/master/graph/badge.svg)](https://codecov.io/gh/bstick12/git-ssh-buildpack)

## Usage

Export your SSH key to an environment variable named `GIT_SSH_KEY`

```shell
export GIT_SSH_KEY=`cat /path/to/ssh_key`
```

If your private SSH key is stored encrypted on disk, you can decrypt it like
so (it requires the [openssl binary](https://www.openssl.org/)):

```shell
openssl rsa -in ~/.ssh/id_rsa -out ~/.ssh/id_rsa.key
chmod 600 ~/.ssh/id_rsa.key
```

This will store your key on disk in an unencrypted format (PEM). If you don't
want to do that, you can decrypt the password on each "export":

```shell
export GIT_SSH_KEY=`cat ~/.ssh/id_rsa | openssl rsa`
```

Add the buildpack e.g via the pack cli

```shell
# see above to export your key into $GIT_SSH_KEY
cd ~/workspace
git clone https://github.com/bstick12/git-ssh-buildpack.git
cd <your project>
pack build <image-name> --builder cloudfoundry/cnb:cflinuxfs3 --buildpack ~/workspace/git-ssh-buildpack --env GIT_SSH_KEY ....
```

## Development

* `scripts/unit.sh` - Runs unit tests for the buildpack
* `scripts/build.sh` - Builds the buildpack
* `scripts/package.sh` - Package the buildpack

## Authors

* [Brendan Nolan](https://github.com/bstick12)
* [Khaled Blah](https://github.com/khaledavarteq)
