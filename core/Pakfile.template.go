package core

const PakfileTemplate = `# Pakfile[.yaml]
# Usage:
#
# Write down your packages dependences that you want pak to manage.
# You can specify repository and branches which also has some default
# value, according to different version control tool.
# The option comes with a preceding '@' and follows after the package name.
# Master branch is the default branch to be checked out if no branch is
# specified.
#
# Examples:
#
# packages:
# - github.com/user/project1
# - github.com/user/project1@dev
#

packages:
- name:
  pakname: pak
  targetbranch: origin/master
`
