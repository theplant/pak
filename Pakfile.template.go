package pak

const PakfileTemplate = `# Pakfile[.yaml]
# Usage:
#
# Write down your package name to be checked out. You can specify branch.
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
# - package@dev`
