# duplicacy-util: Schedule and run duplicacy via CLI

[![Go Report Card](https://goreportcard.com/badge/github.com/jeffaco/duplicacy-util)](https://goreportcard.com/report/github.com/jeffaco/duplicacy-util)
[![Build Status](https://travis-ci.org/jeffaco/duplicacy-util.svg?branch=master)](https://travis-ci.org/jeffaco/duplicacy-util)

[duplicacy]: https://github.com/gilbertchen/duplicacy
[viper]: https://github.com/spf13/viper

This repository contains utilities to run [Duplicacy][] on any platform supported by
[Duplicacy][].

Table of contents:

- [What is duplicacy-util?](#what-is-duplicacy-util)
- [Build instructions](#build-instructions)
- [How do you configure duplicacy-util?](#how-do-you-configure-duplicacy-util)
  - [Global configuration file](#global-configuration-file)
    - [Notifications](#notifications)
    - [E-Mail Notifications](#email-notifications)
  - [Local configuration file](#local-configuration-file)
- [Command line usage](#command-line-usage)
- [Getting started with duplicacy-util](#getting-started-with-duplicacy-util)
- [Management of E-Mail Messages](#management-of-e-mail-messages)
- [Scheduling duplicacy-util to run automatically](#scheduling-duplicacy-util-to-run-automatically)
  - [Scheduling for Linux](#scheduling-for-linux)
  - [Scheduling for Mac OS/X](#scheduling-for-mac-osx)
  - [Scheduling for Windows](#scheduling-for-windows)

---

### What is duplicacy-util?

In short, `duplicacy-util` is a utility to run Duplicacy backups. While
there are a number of
[other tools](https://github.com/gilbertchen/duplicacy/wiki/Scripts-and-utilities-index)
available to do similar things, `duplicacy-util` has a number of advantages:

- It is completely portable. It's trivial to run on Windows, Mac OS/X,
  Linux, or any other platorm supported by the Go language. You schedule
  it, and `duplicacy-util` will perform the backups. Note that [Duplicacy][]
  itself is written in Go, so if you can use [Duplicacy][], you can use
  `dupliacy-util`.
- It is self-contained. Copy a single executable, and `duplicacy-util` is
  fully functional. It is easy to install and easy to upgrade, and you don't
  need to install packages to make it work properly.
- It is "set and forget". I use `duplicacy-util` to send E-Mail upon completion. Then I
  run scripting on my E-Mail server (I use gmail) to move successful backups to the trash.
  This means that I can review backups at any time but, if I don't, the mail messages are
  deleted after 30 days. If any backup fails, it's left in your inbox for you to review.
  See [management of E-Mail messages](#management-of-e-mail-messages) for details.
- It is completely configurable with configuration files. You can have one backup that
  is backed up to a single server while other backups are backed up to multiple servers.
- It is designed to be easy on resources. For example, any number of complete logs are
  saved, but older logs are compressed to save space. Very old logs are aged out and
  deleted.
- `duplicacy-util` won't step on itself. You can run multiple backups concurrently, but
  `duplicacy-util` will skip a backup if it's already backing up a specific repository. Thus, you
  can schedule jobs as often as you would like knowing that if a backup of a repository
  is still running, a second job won't try to back up the same data again.

Note that `duplicacy-util` is a work in progress. The short term to-do list includes:

- Create a checkpoint mechnanism. If [Duplicacy][] fails for whatever reason, then
  `duplicacy-util` should resume the backup where it left off, even if you back up to
  many different storages.
- While designed for my usage, I would very much like feedback to see what others would
  like. If a new feature makes sense, I'm happy to add it.

### Build Instructions

Note that binaries for common platforms are provided. See
[releases on GitHub](https://github.com/jeffaco/duplicacy-util/releases)
for the distributions. However, if you wish to build `duplicacy-util`
yourself, follow instructions in this section.

Building `duplicacy-util` from source is easy. First
[install Go](https://golang.org/doc/install) itself. Once Go is installed
and `$GOPATH` is set up, run the following commands from the command line
to get dependencies:

```shell
go get github.com/djherbis/times
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

### How do you configure duplicacy-util?

`duplicacy-util` works off of two (or more) configuration files:

- A global configuration file (that controls common settings), and
- A repository-specific file to control how the repository should be backed up.

You can have multiple repository-specific configuration files (if you have
many repositories to back up).

Configuration file formats are very flexible. Configuration files can be in
JSON, TOML, YAML, HCL, or Java properties config files (configuration files
are managed with [Viper][]). All examples for configuration files will be in
YAML, but you are to free to use a format of your choosing.

Note that the extension of configuration files can vary based on the format
of the file. Sample configuration files are YAML files, and thus have a YAML
extension. Change the extension if you wish to use JSON or some other format.

By default, dupliacy-util stores all files in its _storage directory_, which is
`$HOME/.duplicacy-util` by default. Note that, in this document, `$HOME` refers
to the users home directory (`~/` on Mac OS/X and Linux, or `/Users/<username>`
on Windows).

The storage directory is determined in a variety of ways:

1.  First and foremost, if the `-sd` parameter is specified, this will define
    the location of the storage directory, and `duplicacy-util` files will be stored
    directly in this directory. In this way, the directory where `duplicacy-util` stores
    its files could be called anything.

1.  If `-sd` is not specified on the command line, then the value of environment
    variable "$HOME" will be evaluated and will be used as a location to look for
    directory `.duplicacy-util`.

1.  If environment variable "$HOME" is unmodified (or not normally defined on your
    system), then it is expected that directory `.duplicacy-util` exists in the users
    home directory.

#### Global configuration file

The global configuration file is called `duplicacy-util.yaml`, and is searched
in the _storage directory_.

The following fields are checked in the global configuration file:

| Field Name          | Purpose                                              | Default Value                                   |
| ------------------- | ---------------------------------------------------- | ----------------------------------------------- |
| duplicacypath       | Path for the [Duplicacy] binary program              | "duplicacy" on your default path ($PATH)        |
| lockdirectory       | Directory where temporary lock files are stored      | Storage directory, or $HOME/.duplicacy-util     |
| logdirectory        | Directory where log files are stored                 | Storage directory, or $HOME/.duplicacy-util/log |
| logfilecount        | Number of historical log files that should be stored | 5                                               |

##### Notifications

`Duplicacy-util` supports notifying you when backups start, are skipped (if
already running), succeed, and fail. Unless you're planning to only be running
`dupliacy-util` interactively, it's strongly recommended to configure
notifications.

For now only email notifications are supported, but more notification channels
will be implemented. The following config snippet shows how to subscribe to
specific notifications:

```
notifications:
  onStart: []
  onSkip: ['email']
  onSuccess: ['email']
  onFailure: ['email']
```

##### Email notifications

| Field Name          | Purpose                                              | Default |
| ------------------- | ---------------------------------------------------- | ------- |
| fromAddress         | From address (i.e. from-user@domain.com)             | None    |
| toAddress           | To address (i.e. to-user@domain.com)                 | None    |
| serverHostname      | SMTP server (i.e. smtp@gmail.com)                    | None    |
| serverPort          | Port of SMTP server (i.e. 465 or 587)                | None    |
| authUsername        | Username for authentication with SMTP server         | None    |
| authPassword        | Password for authentication with SMTP server         | None    |
| acceptInsecureCerts | Accept insecure or self-signed server certificates   | false   |

Notes on email fields:
* If you don't wish to store your email authentication password in the global
  configuration file, you can set environment variable `DU_EMAIL_AUTH_PASSWORD`
  to your email server password. If this environment variable is not defined,
  then we'll check the global configuration file for the password.
* If you are using a local email server, you are likely using a self-signed
  certificate. If that's the case, you should set `acceptInsecureCerts` to
  `true` so `duplicacy-util` won't reject the server certificate.

Here is an example how to setup email notifications:

```
notifications:
  onStart: []
  onSkip: ['email']
  onSuccess: ['email']
  onFailure: ['email']

email:
  fromAddress: "Donald Duck <donald.xyzzy@gmail.com>"
  toAddress: "Donald Duck <donald.xyzzy@gmail.com>"
  serverHostname: smtp.gmail.com
  serverPort: 465
  authUsername: donald.xyzzy@gmail.com
  authPassword: gaozqlwbztypagwt
```

E-Mail subjects from `duplicacy-util` will be of the following format:

| Notification    | Subject Line                                                               |
| --------------- | -------------------------------------------------------------------------- |
| Start           | `duplicacy-util: Backup started for configuration <config-name>`           |
| Skip            | `duplicacy-util: Backup results for configuration <config-name> (skipped)` |
| Success         | `duplicacy-util: Backup results for configuration <config-name> (success)` |
| Failure         | `duplicacy-util: Backup results for configuration <config-name> (FAILURE)` |

You can filter on the subject line to direct the E-Mail appropriately
to a folder of your choice.
See [Management of E-Mail Messages](#management-of-e-mail-messages), for E-Mail configuration hints.

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

- You have a repository, stored in /Volumes/Quicken,
- That is backed up to storage named `b2`,
- You should prune storage `b2` with `0:365 30:180 7:30 1:7`. See
  [prune documentation](https://github.com/gilbertchen/duplicacy/wiki/prune)
  for more information on how to specify `keep` tag.
- When doing a `check` operation, you should check revisions in storage
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

| Field Name | Purpose                               | Default Value |
| ---------- | ------------------------------------- | ------------- |
| repository | Location of the repository to back up | None          |

Sections in the repository configuration files consist of:

| Section Name | Purpose                                                                                                                |
| ------------ | ---------------------------------------------------------------------------------------------------------------------- |
| storage      | Storage names to back up for [duplicacy backup](https://github.com/gilbertchen/duplicacy/wiki/backup) operations\*     |
| copy         | List of storage from-to pairs for [duplicacy copy](https://github.com/gilbertchen/duplicacy/wiki/copy) operations      |
| prune        | List of storage names to prune for [duplicacy prune](https://github.com/gilbertchen/duplicacy/wiki/prune) operations\* |
| check        | List of storage names to check for [duplicacy check](https://github.com/gilbertchen/duplicacy/wiki/check) operations\* |

Note that `*` denotes that this section is mandatory and MUST be specified
in the configuration file.

The `storage` list contains a list of repositories to back up. Note that the
list may be as long as required. `duplicacy-util` will continue loading
monotonically increasing section numbers until no additional sections are
found (i.e. 1, 2, 3). This is conistent with all sections in the repository
configuration file.

Fields in the `storage` section are:

| Field Name | Purpose                                                                         | Required | Default Value |
| ---------- | ------------------------------------------------------------------------------- | -------- | ------------- |
| name       | Storage name to back up                                                         | Yes      | None          |
| threads    | Number of threads to use for backup                                             | No       | 1             |
| vss        | Enable Volume Shadow Copy service                                               | No       | false         |
| vssTimeout | the timeout in seconds to wait for the Volume Shadow Copy operation to complete | No       | None          |

Fields in the `copy` section (if one exists), are:

| Field Name | Purpose                           | Required | Default Value |
| ---------- | --------------------------------- | -------- | ------------- |
| from       | Storage name to copy from         | Yes      | None          |
| to         | Storage name to copy to           | Yes      | None          |
| threads    | Number of threads to use for copy | No       | 1             |

Fields in the `prune` section are:

| Field Name | Purpose                                                                        | Required | Default Value |
| ---------- | ------------------------------------------------------------------------------ | -------- | ------------- |
| storage    | Storage name to prune                                                          | Yes      | None          |
| keep       | [Retention specification](https://github.com/gilbertchen/duplicacy/wiki/prune) | Yes      | None          |

Finally, fields in the `check` section are:

| Field Name | Purpose                         | Required | Default Value |
| ---------- | ------------------------------- | -------- | ------------- |
| storage    | Storage name to check           | Yes      | None          |
| all        | Should all revisions be checked | No       | false         |

Once you have the configuration files set up, running `duplicacy-util` is
simple. Just use a command like:

`duplicacy-util -f quicken -a`

This says: Back up repository defined in `quicken.yaml`, performing all
operations (back up/copy, prune, and check).

Output from this command is similar to:

```text
17:58:25 Using global config: /Users/jeff/.duplicacy-util/duplicacy-util.yaml
17:58:25 Using config file:   /Users/jeff/.duplicacy-util/quicken.yaml
17:58:25 duplicacy-util starting, version: <dev>, Git Hash: <unknown>
17:58:25 Rotating log files
17:58:25 Beginning backup on 07-17-2018 17:58:25
17:58:25 Backing up to storage b2 with 10 threads
17:58:32   Files: 345 total, 823,165K bytes; 1 new, 7,964K bytes
17:58:32   All chunks: 150 total, 890,186K bytes; 5 new, 8,086K bytes, 3,092K bytes uploaded
17:58:32   Duration: 7 seconds
17:58:32 Backing up to storage azure-direct with 5 threads
17:58:33   Files: 345 total, 823,165K bytes; 1 new, 7,964K bytes
17:58:33   All chunks: 150 total, 889,922K bytes; 5 new, 8,086K bytes, 3,092K bytes uploaded
17:58:33   Duration: 1 second
17:58:33 Copying from storage b2 to storage azure with 10 threads
17:58:37   Copy complete, 110 total chunks, 3 chunks copied, 107 skipped
17:58:37   Duration: 4 seconds
17:58:37 Pruning storage b2
17:58:44 Pruning storage azure
17:58:45 Checking storage b2
17:58:47 Checking storage azure
17:58:48 Operations completed in 23 seconds
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
  -a	Perform all duplicacy operations (backup/copy, purge, check)
  -b	Perform duplicacy backup/copy operation
  -c	Perform duplicacy check operation
  -d	Enable debug output (implies verbose)
  -f string
    	Configuration file for storage definitions (must be specified)
  -g string
    	Global configuration file name
  -m	(Deprecated) Send E-Mail with results of operations (implies quiet)
  -p	Perform duplicacy prune operation
  -q	Quiet operations (generate output only in case of error)
  -sd string
    	Full path to storage directory for configuration/log files
  -tm
    	(Deprecated: Use -tn instead) Send a test message via E-Mail
  -tn
    	Test notifications
  -v	Enable verbose output
  -version
    	Display version number
```

Exit codes from `duplicacy-util` are as follows:

| Exit Code/Range | Meaning                                         |
| --------------- | ----------------------------------------------- |
| 0               | Success                                         |
| 1-2             | Command line errors                             |
| 500             | Operation from `duplicacy` command failed       |
| 6200            | Run skipped due to existing job already running |

In the event of an error, a notification will be sent with details of the
error. Note that 200-201 operations are not considered fatal from an notification
perspective, but the fact that the backup was skipped is indicated.

### Getting started with duplicacy-util

The `duplicacy-util` program has no knowledge of [Duplicacy][] repository passwords.
As a result, if [Duplicacy][] prompts for a password, `duplicacy-util` won't be able
to respond to the prompt, and the backup will fail (with suitable output in the log
file).

To set up the backup for initial use, there is
[documentation](https://github.com/mattjm/duplicacy-script) that @mattjm worked up
that is pretty good. That said, these are the basic steps I followed to initialize
backing up Quicken, one of my repositories:

```
duplicacy init -e -storage-name b2 quicken b2://<bucket-name>
duplicacy add -e -copy b2 azure quicken azure://<bucket-name>      # Copy
duplicacy add -e azure-direct quicken-direct azure:<bucket-name>   # Direct

duplicacy backup -storage b2 -stats -threads 10
duplicacy backup -storage azure-direct -stats -threads 5
duplicacy copy -from b2 -to azure -stats -threads 10
```

This initialized the repository and set it up for backup to both Backblaze and
Azure. It also performed the first backup, taking care of final password prompts.
After this, `duplicacy-util` should function properly, and [Duplicacy][] should
not prompt for passwords.

You should study the [Duplicacy Wiki](https://github.com/gilbertchen/duplicacy/wiki)
carefully, the documentation is quite good. It explains how [Duplicacy][] works
and various commands that [Duplicacy][] supports.

### Management of E-Mail Messages

**NOTE: This discussion is specific to Gmail, but if you are using a
different mail server, you can almost certainly use these ideas in your
specific scenerio.**

In order to send E-Mail notifications, you must first have configured a 
number of fields in the [Global configuration file](#global-configuration-file).
These fields depend on what E-Mail server
you are using. I use Google's
[gmail](https://en.wikipedia.org/wiki/Gmail)
service, and will define my usage here.


It is recommended that you use an application specific generated
password that can be generated in the
[Gmail Security Center](https://myaccount.google.com/security).
This works around two-factor authentication or other issues that may
create problems. Note that the password stored in the global configuration
file is not encrypted at this time. On a shared system, you should set
permissions of this file appropriately, or use environment variable
`DU_EMAIL_AUTH_PASSWORD` to override the value stored in the global
configuration file.

Once you set up the E-Mail configuration appropriately, you can test it
with a command like: `./duplicacy-util -tn`. This will trigger a failure 
notification for all configured notification channels (e.g E-Mail).

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
  // Check to see that it starts with "duplicacy-util..." and ends with " (Success)", and if so, count
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
    GmailApp.sendEmail('<user>@<domain>.com',
                       'WARNING: No duplicacy-util backup logs files received',
                       'Please investigate backup process!');
  }
}
```

**Be certain to replace `<user>` with your Gmail username and `<domain>`
with your Gmail domain in the script above.**

After the script is set up, you can set up Google to run the script
automatically on any schedule you wish.

### Scheduling duplicacy-util to Run Automatically

Scheduling `duplicacy-util` to run backups automatically (emailing the
results automatically) finishes the job. Now backups run attended,
automatically, relieving you of the job of doing backups yourself.

Backup scheduling differs by operating system. I provide hints here,
although there are lots of diferent ways to schedule jobs automatically.

#### Scheduling for Linux

Linux has a built-in rich scheduler, `cron`. The `cron` utility can run
jobs as a user or as root; the choice is yours. These instructions assume
that you will be running jobs as your user since you'll generally be
backing up your user files.

There's a lot of help available for `cron`.
[Wikipedia](https://en.wikipedia.org/wiki/Cron)
help is a good start for the average user. For purposes of example,
you can do something like the following:

```bash
crontab -l > crontab
echo "0 1 * * * /Users/jeff/Applications/duplicacy-util -f quicken -a -m -q" >> crontab
crontab < crontab
```

The first command will dump your existing crontab entries to a file named
`crontab`. This file will likely be empty if you haven't used `crontab`
before.

The second command will add an entry to your `crontab` file:
Run `duplicacy-util` for Quicken, e-mailing results, at 1:00 AM every
morning. See [Wikipedia](https://en.wikipedia.org/wiki/Cron) for help
in understanding the time format.

Since crontab stores entries internally, the final command will reload
your saved crontab entries from your private `crontab` file.

#### Scheduling for Mac OS/X

On recent versions of Mac OS/X (_macOS High Sierra_ as of the time of
this writing), [cron](https://en.wikipedia.org/wiki/Cron) ships with
Mac OS/X. So that is an option.

[launchd]: https://developer.apple.com/library/content/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/ScheduledJobs.html
[lingon]: https://www.peterborgapps.com/lingon/

However, on Mac OS/X, the preferred way to add a timed job is to use
[launchd][]. Each [launchd][] job is described by a separate file.
This means that you can manage launchd timed jobs by simply adding
or removing a file.

There are two ways to create these files:

1.  By hand; the file format is documented in [launchd][] documentation, or
1.  By using an automated tool. [Lingon][] is one such tool that makes
    the job of creating [launchd][] files very simple. While [Lingon][] is
    commercial, it's very inexpensive. I created the `quicken` job in
    seconds using [Lingon][]:

```
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
        <key>EnvironmentVariables</key>
        <dict>
                <key>PATH</key>
                <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin:/usr/local/go/bin:/opt/X11/bin:/usr/local/sbin</string>
        </dict>
        <key>Label</key>
        <string>com.duplicacy-util.quicken</string>
        <key>ProgramArguments</key>
        <array>
                <string>/Users/jeff/local/go/bin/duplicacy-util</string>
                <string>-f</string>
                <string>quicken</string>
                <string>-a</string>
                <string>-m</string>
                <string>-q</string>
        </array>
        <key>RunAtLoad</key>
        <false/>
        <key>StartCalendarInterval</key>
        <array>
                <dict>
                        <key>Hour</key>
                        <integer>3</integer>
                        <key>Minute</key>
                        <integer>0</integer>
                </dict>
                <dict>
                        <key>Hour</key>
                        <integer>15</integer>
                        <key>Minute</key>
                        <integer>0</integer>
                </dict>
        </array>
</dict>
</plist>
```

This `plist` file will run job `quicken` twice a day: at 3:00 AM and
at 3:00 PM, mailing the results of the backup job.

#### Scheduling for Windows

Windows includes a build-in rich scheduler called the `Windows Task Scheduler`. The `Windows Task Scheduler` is a GUI (graphical) program
designed to make scheduling of repetitive tasks easy to perform.

You can find help in numerous forms with a
[WWW search](https://www.google.com/search?q=how+to+use+windows+task+scheduler),
including articles and YouTube videos stepping you through the process.
