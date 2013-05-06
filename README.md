# Pak [![Build Status](https://travis-ci.org/theplant/pak.png?branch=master)](https://travis-ci.org/theplant/pak)

Whether Pak is useful or not depends on what kind of the development strategy that you and your team take. When you are developing a project, which is divided into few smaller projects, like bellow:

	github.com/team/project
	github.com/team/sub-project2
	github.com/team/sub-project3

And in each project, you use different branches or something to mark project state. For instance, you might define that in each project, a branch named master is used for production, and branch dev used for implementing new features. Sooner or later, you will find out that you have to switch branches between projects from time to time. And when you switch `github.com/team/project` from branch master to branch dev, you may also need to make sure `sub-project2` and `sub-project3` is also on the branch dev. When you update some of them, you may also need to remember to remind your teammates of the updates.

Experiences like this is nuisance. So Pak comes.

## Introduction

Pak uses a Pakfile and Pakfile.lock for easy package version synchronisation and management. After specify packages in Pakfile, running `get` and `update` command to ask Pak get your dependencies on the right state.

For example, If you have a Pakfile like bellow:

```
packages:
- github.com/theplant/package1				# custom branch master and default remote origin
- github.com/theplant/package2@dev			# custom branch and default remote origin
- github.com/theplant/package3@origin/dev	# custom remote and branch
```

The first time you run `pak get`, Pak, under the instruction of `Pakfile`, will check out a branch named `pak` from branch `refs/remotes/origin/master` in `package1`, similar to package2 and package3.

After that, it will generate a `Pakfile.lock` file, it is necessary to check it into your version control system, and when you teammates run `pak get`, Pak will not use `Pakfile` anymore, `Pakfile.lock` will be used to get those packages on the same state with you.

If someone has submitted new changes into those decencies, you want to check those changes in, use `pak update`. It will fetches the latest changes from remote repo first, and then checks out the latest the changes and updates `Pakfile.lock`.

Pak borrows a lot of concepts from [Bundler](http://gembundler.com/).

## Installation & Usage

Installation is simple.

```
go get -u github.com/theplant/pak
```

After installation, use `pak` to check the help messages out.

Currently available commands are list bellow:

```
Usage:
    pak init
    pak [-sf] get [package]
    pak [-s] update [package]
    pak open [package]
    pak list
    pak version
  -f=false: Force pak to remove pak branch.
  -s=false: Left out unclean packages.
```

### Auto-Checking

This feature is used to force your app dependencies to be up-to-date with Pakfile and Pakfile.lock.

In your package, use it as bellow:

```go
import "github.com/theplant/pak/check"

func init() {
    check.Check()
}
```

And each time you start your app, pak will auto check the dependencies of your app. if your app is not consistent with Pakfile and Pakfile.lock, it will force your app to exit. Like this:

![Check](https://raw.github.com/theplant/pak/master/imgs/check.png?login=bom-d-van&token=93b3b310df07f7163a3b57efe9fa0ada)

## Others

Currently, git and mercurial are supported.

<!--## License

Train is released under the [MIT License](http://www.opensource.org/licenses/MIT).-->
