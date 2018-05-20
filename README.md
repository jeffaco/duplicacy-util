# duplicacy-util: Schedule and run duplicacy via CLI

[Duplicacy]: https://github.com/gilbertchen/duplicacy
[Viper]: https://github.com/spf13/viper

This repository contains utilities to run [Duplicacy][] on any platform supported by
[Duplicacy][].

Table of contents:

* [What is duplicacy-util?](#what-is-duplicacy-util)
* [Build Instructions](#build-instructions)
* [How do you configure duplicacy-util?](#how-do-you-configure-duplicacy-util)
  * [Global configuration file](#global-configuration-file)
  * [Local configuration file](#local-configuration-file)
* [Command line usage](#command-line-usage)
* [Management of E-Mail Messages](#management-of-e-mail-messages)

-----

### What is duplicacy-util?

In short, `duplicacy-util` is a utility to run Duplicacy backups. While
there are a number of
[other tools](https://github.com/gilbertchen/duplicacy/wiki/Scripts-and-utilities-index)
available to do similar things, duplicacy-util has a number of advantages:

* It is completely portable. It's trivial to run on Windows, Mac OS/X,
Linux, or any other platorm supported by the Go language. You schedule
it, and duplicacy-utils will perform the backups. Note that [Duplicacy][]
itself is written in Go, so if you can use [Duplicacy][], you can use
dupliacy-util.
* It is self-contained. Copy a single executable, and `duplicacy-util` is
fully functional. It is easy to install and easy to upgrade, and you don't
need to install packages to make it work properly.
* It is "set and forget". I use duplicacy-utils to send E-Mail upon completion. Then I
run scripting on my E-Mail server (I use gmail) to move successful backups to the trash.
This means that I can review backups at any time but, if I don't, the mail messages are
deleted after 30 days. If any backup fails, it's left in your inbox for you to review.
See [management of E-Mail messages](#management-of-e-mail-messages) for details.
* It is completely configurable with configuration files. You can have one backup that
is backed up to a single server while other backups are backed up to multiple servers.
* It is designed to be easy on resources. For example, any number of complete logs are
saved, but older logs are compressed to save space. Very old logs are aged out and
deleted.
* duplicacy-util won't step on itself. You can run multiple backups concurrently, but
duplicacy-util will skip a backup if it's already backing up a specific repository. Thus, you
can schedule jobs as often as you would like knowing that if a backup of a repository
is still running, a second job won't try to back up the same data again.

Note that duplicacy-util is a work in progress. The short term to-do list includes:

* Complete e-mail notification (a framework is in place, but not yet completed). This
will be done very shortly (within days).
* Create a checkpoint mechnanism. If [Duplicacy][] fails for whatever reason, then
duplicacy-util should resume the backup where it left off, even if you back up to
many different storages.
* While designed for my usage, I would very much like feedback to see what others would
like. If a new feature makes sense, I'm happy to add it.

### Build Instructions

Building duplicacy-util from source is easy. First
[install Go](https://golang.org/doc/install) itself. Once Go is installed
and `$GOPATH` is set up, run the following commands from the command line
to get dependencies:

```shell
go get github.com/mitchellh/go-homedir       
go get github.com/spf13/viper
go get github.com/theckman/go-flock
go get gopkg.in/gomail.v2
```

Finally, download `duplicacy-util` itself:

```shell
cd $GOPATH/src
git clone https://github.com/jeffaco/duplicacy-util.git
```

Once Go is installed and dependencies are downloaded, to build, do:

```shell
cd $GOPATH/src/duplicacy-util
go build
```

This will generate a `duplicacy-util` binary in the current directory with
the appropriate file extension for your platform (i.e. `duplicacy-util` for
Mac OS/X or Linux, or `duplicacy-util.exe` for Windows).

Note that once development of `duplicacy-util` is more complete, I'll post
binaries for common platforms. See
[releases on GitHub](https://github.com/jeffaco/duplicacy-util/releases).

### How do you configure duplicacy-util?

duplicacy-util works off of two (or more) configuration files:

* A global configuration file (that controls common settings), and
* A repository-specific file to control how the repository should be backed up.

You can have multiple repository-specific configuration files (if you have
many repositories to back up).

Configuration file formats are very flexible. Configuration files can be in
JSON, TOML, YAML, HCL, or Java properties config files (configuration files
are managed with [Viper][]). All examples for configuration files will be in
YAML, but you are to free to use a format of your choosing.

Note that the extension of configuration files can vary based on the format
of the file. Sample configuration files are YAML files, and thus have a YAML
extension. Change the extension if you wish to use JSON or some other format.

By default, dupliacy-util stores all files in `$HOME/.duplicacy-util`. This
can be changed via the global configuration file. Note that, in this document,
`$HOME` refers to the users home directory (`~/` on Mac OS/X and Linux, or
`/Users/<username>` on Windows).

#### Global configuration file

The global configuration file is called `duplicacy-util.yaml`. We search in
`$HOME` and `$HOME/.duplicacy-util` looking for this file.

The following fields are checked in the global configuration file:

Field Name | Purpose | Default Value
---------- | ------- | -------------
duplicacypath | Path for the [Duplicacy] binary program | "duplicacy" on your default path ($PATH)
lockdirectory | Directory where temporary lock files are stored | $HOME/.duplicacy-util
logdirectory | Directory where log files are stored | $HOME/.duplicacy-util/log
logfilecount | Number of historical log files that should be stored | 5

A sample global configuration file is stored in `~/.duplicacy-util/duplicacy-util.yaml`,
and contains the following:

```
duplicacypath: /Users/Jeff/Applications/duplicacy
```

#### Local configuration file

The local configuration file (or repository configuration file) defines how to
back up a specific repository. This file must be specified on the command line
(discussed later). The repository-specific configuration file may take lists
of storages if you
[back up to multiple cloud providers](https://github.com/gilbertchen/duplicacy/wiki/Back-up-to-multiple-storages).
In the simple case, a configuration file can short, such as this:

```
repository: /Volumes/Quicken

storage:
    1:
        name: b2

prune:
    1:
        storage: b2
        keep: "0:365 30:180 7:30 1:7"

check:
    1:
        storage: b2
```

This configuration shows that:

* You have a repository, stored in /Volumes/Quicken,
* That is backed up to storage named `b2`,
* You should prune storage `b2` with `0:365 30:180 7:30 1:7`. See
[prune documentation](https://github.com/gilbertchen/duplicacy/wiki/prune)
for more information on how to specify `keep` tag.
* When doing a `check` operation, you should check revisions in storage
`b2`.

You might wonder why the same storage is specified multiple times. This is
evident if you back up to multiple cloud providers.

If you back up to multiple cloud providers, the configuration file may be
more involved:

```
repository: /Volumes/Quicken

storage:
    1:
        name: b2
        threads: 10
    2:
        name: azure-direct
        threads: 5

copy:
    1:
        from: b2
        to: azure
        threads: 10

prune:
    1:
        storage: b2
        keep: "0:365 30:180 7:30 1:7"
    2:
        storage: azure
        keep: "0:365 30:180 7:30 1:7"

check:
    1:
        storage: b2
        all: true
    2:
        storage: azure
        all: true
```

The new concept here is the `copy` section. This defines repositories that
should be copied from one storage to another, but using a pseudo storage
name (`azure-direct`) to avoid downloading a lot of data from `b2`. In
this example, we'll back up to both `b2` and `azure-direct`, but then we'll
use a `duplicacy copy` operation to be sure that the two storages are
identical when the backup is complete.

Because there are multiple storages involved, we want to prune each storage
and check each storage for consistency.

A repository configuration file consists of a few repository-wide settings
and sections that define operations. The repository-wide settings are:

Field Name | Purpose | Default Value
---------- | ------- | -------------
repository | Location of the repository to back up | None

Sections in the repository configuration files consist of:

Section Name | Purpose
------------ | -------
storage | Storage names to back up for [duplicacy backup](https://github.com/gilbertchen/duplicacy/wiki/backup) operations*
copy | List of storage from-to pairs for [duplicacy copy](https://github.com/gilbertchen/duplicacy/wiki/copy) operations
prune | List of storage names to prune for [duplicacy prune](https://github.com/gilbertchen/duplicacy/wiki/prune) operations*
check | List of storage names to check for [duplicacy check](https://github.com/gilbertchen/duplicacy/wiki/check) operations*

Note that `*` denotes that this section is mandatory and MUST be specified
in the configuration file.

The `storage` list contains a list of repositories to back up. Note that the
list may be as long as required. duplicacy-util will continue loading
monotonically increasing section numbers until no additional sections are
found (i.e. 1, 2, 3). This is conistent with all sections in the repository
configuration file.

Fields in the `storage` section are:

Field Name | Purpose | Required | Default Value
---------- | ------- | -------- | -------------
name | Storage name to back up | Yes | None
threads | Number of threads to use for backup | No | 1

Fields in the `copy` section (if one exists), are:

Field Name | Purpose | Required | Default Value
---------- | ------- | -------- | -------------
from | Storage name to copy from | Yes | None
to | Storage name to copy to | Yes | None
threads | Number of threads to use for copy | No | 1

Fields in the `prune` section are:

Field Name | Purpose | Required | Default Value
---------- | ------- | -------- | -------------
storage | Storage name to prune | Yes | None
keep | [Retention specification](https://github.com/gilbertchen/duplicacy/wiki/prune) | Yes | None

Finally, fields in the `check` section are:

Field Name | Purpose | Required | Default Value
---------- | ------- | -------- | -------------
storage | Storage name to check | Yes | None
all | Should all revisions be checked | No | false

Once you have the configuration files set up, running `duplicacy-util` is
simple. Just use a command like:

`duplicacy-util -f quicken -a`

This says: Back up repository defined in `quicken.yaml`, performing all
operations (back up/copy, prune, and check). If E-Mail is configured in
the global configuration file, then E-Mail will be sent at the end of
the run.

Output from this command is similar to:

```text
Using global config: /Users/jeff/.duplicacy-util/duplicacy-util.yaml
Using config file:   /Users/jeff/.duplicacy-util/quicken.yaml
17:50:21 Rotating log files
17:50:21 Beginning backup on 05-18-2018 17:50:21
17:50:21 Backing up to storage b2 with 10 threads
17:50:28 Backing up to storage azure-direct with 5 threads
17:50:29 Copying from storage b2 to storage azure with 10 threads
17:50:33 Pruning storage b2
17:51:01 Pruning storage azure
17:51:04 Checking storage b2
17:51:06 Checking storage azure
17:51:07 Operations completed in 45.525075541s
```

A complete log of the backup is saved in the `logdirectory` setting in the
global configuration file.

### Command Line Usage

The best way to get command line usage is to run `duplicacy-util` with the
`-h` option, as follows:

`duplicacy-util -h`

This will generate output similar to:

```text
Usage of ./duplicacy-util:
  -a    Perform all duplicacy operations (backup/copy, purge, check)
  -b    Perform duplicacy backup/copy operation
  -c    Perform duplicacy check operation
  -d    Enable debug output (implies verbose)
  -f string
        Configuration file for storage definitions (must be specified)
  -g string
        Global configuration file name
  -m    Send a test message via E-Mail
  -p    Perform duplicacy prune operation
  -v    Enable verbose output
```

Exit codes from duplicacy-util are as follows:

Exit Code/Range | Meaning
--------------- | -------
0 | Success
1-2 | Command line errors
200-201 | Run skipped due to existing job already running
500 | Operation from `duplicacy` command failed

In the event of an error, an E-Mail will be generated with details of the
error. Note that 200-201 operations are not considered fatal from an E-Mail
perspective, but the fact that the backup was skipped is indicated.

E-Mail subjects from duplicacy-util will be of the following format:

Success/Failure | Subject Line
--------------- | ------------
Success | `duplicacy-util backup of quicken results [success]`
Failure | `duplicacy-util backup of quicken results [failed!]`

You can can filter on the subject line to direct the E-Mail appropriately
to a folder of your choice.

### Management of E-Mail Messages

This section will be completed once E-Mail is fully implemented.
