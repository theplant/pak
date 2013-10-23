# Pak

A go packages version control tool, inspired by Bundler for Ruby.

## Quick Introduction

* [Common Usage](http://ascii.io/a/5454)
* [Partial Matching](http://ascii.io/a/5455)
* [Pak Don't Hurt Unclean Packages](http://ascii.io/a/5456)

BTW, this screenshots is made weeks ago, and pak has been improving all the time. So the output might be a bit of different now, however, the way pak works is still the same.

Powerd by [ASCII.IO](http://ascii.io/).

## What Pak Can Do

Assume that you are working on two projects called pro1 and pro2, and they are both depending on two other projects working by your colleagues, req1 and req2.

1. Req1 has two branches, branch1 and branch2. Req2 also has two branches, branch1 and branch2.
2. Pro1 depends on branch1 in req1 and branch1 in req2, but pro2 depends on branch2 in req1 and branch2 in req2.
3. Make it simple, supposed they are all using Git as their version control tool.
4. Sometimes you are working on pro1, sometimes pro2. How to make yourself efficient and productive. Using one GOPATH you have to checkout stuff from req1 and req2. Using two GOPATH you have to switch GOPATH from time to time (may be there are better solution, but that's all I know. :-P).

If you use Pak. The workflow will be like this:

Make Pakfiles in pro1 and pro2 like bellow:

In Pro1:

```yaml
packages:
- name: github.com/theplant/req1
  targetbranch: origin/branch1
- name: github.com/theplant/req2
  targetbranch: origin/branch1
```
In Pro2:

```yaml
packages:
- name: github.com/theplant/req1
  targetbranch: origin/branch2
- name: github.com/theplant/req2
  targetbranch: origin/branch2
```

Then, when you are working on Pro1, you can did this:

```
pak get
```

This action will generates a file named `Pakfile.lock` in the first time that you use `pak`. It will retrieve the up-to-date checksum of packages specified in `Pakfile`. The next time you use `pak get`, pak will try to checkout req1 and req2 using the checksum saved in `Pakfile.lock`

```
github.com/theplant/req1: 6dd3a9a0e8349b0421c57c79b8f45d3565a96378
github.com/theplant/req2: 5e1d544059ce1ff74d833da7f0d5a8ca02a82525
```

Just in pro2, you can do the same thing, it will generate a similar `Pakfile.lock`:

```
github.com/theplant/req1: 931b60b175dcfd6afa02d34e13270b8aaa4d0ba2
github.com/theplant/req2: 1d1416e1f8fce75311d2afe5fc391aac84927601
```

When you go to package req1 or req2, you can see these two packages are on status that you specified in `Pakfile` and `Pakfile.lock`.

## How Pak works

The mechanism of `pak` is pretty simple: Checking out the most up-to-date commit from all the dependences according to the description of `Pakfile` and then store the checksum in `Pakfile.lock`. It's like taking a snapshot of the dependences of your project and no matter how many changes is undergoing in them, each time when you need to run your project, just run `pak get` again then you can cancel all those changes (not actually delete them, just checkout those dependences accroding to `Pakfile.lock`). When you need to update some dependences in the project, just use `pak update`, then `pak` will checkout the up-to-date commit for you.

## Usage

Installation is pretty simple:

```
go get github.com/theplant/pak
```

Then done. See avaliable actions:

```
$: pak
Usage:
    pak init
    pak [-sf] get [package]
    pak [-s] update [package]
    pak open [package]
    pak list
    pak version
  -f=false: Force pak to remove pak branch or pak unclean packages.
  -s=false: Skip unclean packages.
```

### Init and Configuration

```
pak init
```

This Command will generate a file named Pakfile in which you can write down dependences that your project needs.

Pakfile is using YAML syntax.

A sample:

```yaml
packages:
- name: github.com/theplant/package1
  pakname: pak
  targetbranch: origin/master
- name: github.com/theplant/package1
  pakname: pak
  targetbranch: origin/master
```

All package requirements should be listed in packages section. Each package contains some descriptions which are explained below:

`name`: the package name.

`pakname`: a name pak used to checkout a branch/bookmark in dependent packages, default value is pak.

`targetbranch`: the branch which you need pak to monitor, default values in git is `origin/master` and `default/default` in mercurial. It must be a remote branch.

Every dependences in Pakfile must have a remote repository. Pak will only try to checkout them from that remote repository. The reason behind this is `pak` is designed to be an cooperating tool first then an package management tool. And a package without remote repository is difficult to share with your teammates and other nice guys on Internet.

### Pak Get

After finishing a Pakfile, using `pak get` to take the first snapshot of your project's dependences. That will generate a Pakfile.lock by retrieving the up-to-date checksum from the remote repositories of your dependences.

Without the exitstence of `Pakfile.lock`, `pak` will try to checkout the up-to-date commit according to descriptions of `Pakfile`. So after the first time that you use `pak get`, pak creates `Pakfile.lock`. With the existence of `Pakfile.lock`, `pak get` will checkout commits recorded in `Pakfile.lock` from your dependences.

`pak get` supports you to get specific packages like this:

```
pak get some-package
```

BTW, pak support partial matching, you don't have to type the whole name of that packages.

### Pak Update

After paking some packages for a while, you might be to update some of the packages, here comes `pak update`.

When running `pak update`, pak will retrieve the up-to-date commits from the remote repository and lock them in `Pakfile.lock`.

Also if you don't want to update all dependences (which is slow and might not be unnecessary in some cases), you can specify need-to-update packages like this:

```
pak update some-package
```

### Pak Check

A feature serve as a reminder in your application to help you detect whether the dependences is consistent with your Pakfile.lock. Save you from debugging problems caused by inconsistency in dependences. When

In your package, use it as bellow:

```go
import pak "github.com/theplant/pak/check"

func init() {
    pak.Check()
}
```

And each time you start your app, pak will auto check the dependencies of your app. if your app is not consistent with Pakfile and Pakfile.lock, it will force your app to exit. Like this:

![Check](https://raw.github.com/theplant/pak/master/imgs/check.png?login=bom-d-van&token=93b3b310df07f7163a3b57efe9fa0ada)

Note: It's recommended to use pak.check in your development environemnt instead of production environment. This's mainly a tool for developers.

## Status

Supported Version Control System: Git, Mercurial.

Features:

* Cross Package Dependences
* Auto Detect Package Dependences (pak.Check)

## License

Pak is released under the [MIT License](http://www.opensource.org/licenses/MIT).
