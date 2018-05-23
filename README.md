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
 | |
emailFromAddress | From address (i.e. from-user@domain.com) | None
emailToAddress | To address (i.e. to-user@domain.com) | None
emailServerHostname | SMTP server (i.e. smtp@gmail.com) | None
emailServerPort | Port of SMTP server (i.e. 465 or 587) | None
emailAuthUsername | Username for authentication with SMTP server | None
emailAuthPassword | Password for authentication with SMTP server | None

A sample global configuration file is stored in `~/.duplicacy-util/duplicacy-util.yaml`,
and contains the following:

```
duplicacypath: /Users/Jeff/Applications/duplicacy
```

If you want E-Mail capabilities (and you almost certainly do unless
you're planning on only running `duplicacy-util` interactively), you
must define prefixes starting with `email` in the list above. For
E-Mail configuration hints, see
[Management of E-Mail Messages](#management-of-e-mail-messages).

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
  -m    Send E-Mail with results of operations (implies quiet)
  -p    Perform duplicacy prune operation
  -q    Quiet operations (generate output only in case of error)
  -tm
        Send a test message via E-Mail
  -v    Enable verbose output
  -version
        Display version number
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

To send E-Mail, you must define a number of fields, and these fields
depend on what E-Mail server you are using. I use Google's
[gmail](https://en.wikipedia.org/wiki/Gmail)
service, and will define my usage here.

**NOTE: This discussion is specific to Gmail, but if you are using a
different mail server, you can almost certainly use these ideas in your
specific scenerio.**

Settings in `$HOME/.duplicacy-util/duplicacy-util.yaml`:

Setting | Suggested value
------- | ---------------
emailFromAddress | Your Gmail address (i.e. jeff@gmail.com)
emailToAdress | Your Gmail address (i.e. jeff@gmail.com)
emailServerHostname | smtp.gmail.com
emailServerPort | 465
emailAuthUsername | Your Gmail address (i.e. jeff@gmail.com)
emailAuthPassword | Your application specific generated password

It is recommended that you use an application specific generated
password that can be generated in the
[Gmail Security Center](https://myaccount.google.com/security).
This works around two-factor authentication or other issues that may
create problems. Note that the password stored in the global configuration
file is not encrypted at this time. On a shared system, you should set
permissions of this file appropriately.

Once you set up the E-Mail configuration appropriately, you can test it
with a command like: `./duplicacy-util -tm`. This will trigger two E-Mail
messages to be generated: one "success" message and one "failure" message.

It's recommended that you use Gmail filtering so that failed backups
are visable in your `inbox` while successful backups are set aside for
deletion. To do this, first create a folder named `Backup Logs`.
After the folder is created, then create a filter rule as follows:

```
Matches: from:(from-user@gmail.com) to:(to-user@taltos.com) subject:(duplicacy-util: Backup results for AND (success))
Do this: Skip Inbox, Mark as read, Apply label "Backup Logs", Never send it to Spam
```

After this is done, generate a mail test and verify that you have a
failed test message in `inbox` and a success test message in `Backup Logs`.

To catch if backups that are not running, and to clean up successful
backups from folder `Backup Logs`, it is recommended that you create a small
[Google Apps Script](https://drive.google.com/drive/search?q=type:script)
to do these actions. In this way, if you do nothing, successful backup
logs are deleted after 30 days automatically, and failures go to your
`inbox`, where you can see them and act upon them.

Here is one such
[Google Apps Script](https://drive.google.com/drive/search?q=type:script)
named `duplicacy-util.gs`:

```
function duplicacy_utils() {
  var threads = GmailApp.search('label:"Backup Logs"');
  var foundBackup = 0;

  // Backups from duplicacy-util with no errors get filtered to label "Backup Logs" via Gmail
  // settings. This makes them easy for us to find and iterate over.
  //
  // Backups are scheduled at least as often as this script runs. Thus, if nothing was run when this
  // script runs, then we get active notification that something is wrong with the backup process.
  //
  // Naming conventions with duplicacy-util are formatted like:
  //   "duplicacy-util: Backup results for configuration test (success)" (for successful backups), or
  //   "duplicacy-util: Backup results for configuration test (FAILURE)" (for failed backups)
  // Check to see that it start with "duplicacy-util..." and ends with " (Success)", and if so, count
  // the message.

  for (var i = 0; i < threads.length; i++) {
    var subject = threads[i].getFirstMessageSubject();
    if (subject.indexOf('duplicacy-util: Backup results for configuration') == 0 && subject.indexOf(' (success)') != -1)
    {
      threads[i].moveToTrash();
      foundBackup++;
    }
  }

  if (foundBackup == 0)
  {
    GmailApp.sendEmail('<user>@<domain>.com', 'WARNING: No duplicacy-util backup logs files received', 'Please investigate backup process!');
  }
}
```

**Be certain to replace `<user>` with your Gmail username and `<domain>`
with your Gmail domain in the script above.**

After the script is set up, you can set up Google to run the script
automatically on any schedule you wish.
