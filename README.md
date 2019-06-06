# GIT SSH key buildpack for Github

Export your SSH key to an environment variable named `GIT_SSH_KEY`

```
export GIT_SSH_KEY=`cat /path/to/ssh_key`
```

Add the buildpack via the pack cli

```
pack build <image-name> --builder cloudfoundry/cnb:cflinuxfs3 --buildpack /path/to/git-ssh-buildpack -e GIT_SSH_KEY ....
```

