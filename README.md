# GIT SSH key buildpack for Github/Bitbucket

This is a [Cloud Native Buildpack V3](https://buildpacks.io/) that enables adding an SSH key to the build process and forces git to use SSH rather than HTTPS for communication with [Github](https://github.com) or [Bitbucket](https://bitbucket.org). It requires [Go v1.13](https://golang.org) or later.

This buildpack is designed to work in collaboration with other buildpacks. It is compatible to API v0.2 and it supports the stacks "org.cloudfoundry.stacks.cflinuxfs3" and "io.buildpacks.stacks.bionic". It has been tested with the [pack CLI](https://github.com/buildpack/pack) v0.5.0.

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
git clone https://github.com/avarteqgmbh/git-ssh-buildpack.git
cd <your project>
pack build <image-name> --builder cloudfoundry/cnb:cflinuxfs3 --buildpack ~/workspace/git-ssh-buildpack --env GIT_SSH_KEY ....
```

## Development

* `scripts/unit.sh` - Runs unit tests for the buildpack
* `scripts/build.sh` - Builds the buildpack
* `scripts/package.sh` - Package the buildpack

## Authors

This project was forked from https://github.com/bstick12/git-ssh-buildpack

* [Brendan Nolan](https://github.com/bstick12)
* [Khaled Blah](https://github.com/khaledavarteq)
