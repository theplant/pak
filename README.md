# Pak

Whether Pak is useful or not depends on the development strategy that you and your teammates take. When you are developing a project, and you divides it into a few smaller projects, like bellow:

	github.com/team/project
	github.com/team/sub-project2
	github.com/team/sub-project3

And in each project, you use different branches for different intensions. For instance, you define that in each project, branch master is used for production, and branch dev is used for implementing new features. Sooner or later, you will find out that you have to switch between different branches from time to time. And when you switch `github.com/team/project` from branch master to branch dev, you also need to make sure `sub-project2` and `sub-project3` is also on the branch dev. And also, you need to remember to check out some new changes committed by your teammates.

Such process could be annoying. So there is Pak.

## Introduction

Pak uses a Pakfile and Pakfile.lock for easy package state management and synchronisation. After you specify packages to be managed by Pak, use `get` and `update` command to ask Pak get your dependencies on the right state.

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
    pak [-uf] get
    pak update [package]

```

## Others

Curretnly, Pak just support git. New commands and other version control system might be supported in the future.


<!--## License

Train is released under the [MIT License](http://www.opensource.org/licenses/MIT).-->