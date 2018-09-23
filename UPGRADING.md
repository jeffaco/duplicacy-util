# UPGRADING.md - Notes on upgrading from prior versions

Table of contents:

- [Explanation of Upgrade Policy](#explanation-of-upgrade-policy)
- [Changes made in `v1.4`(#changes-made-in-v1.4)
  - [Notifications](#v1.4-notifications)
  - [Command Line Changes](#v1.4-command-line-changes]

### Explanation of Upgrade Policy

While reasonable attempts are made to insure compatibility from release
to release, sometimes changes are necessary.

As a policy, when an incompatible change is made, old behavior will be
maintained for two releases (major or minor). So, for example, if
`duplicacy-util v1.4` introduces incompatible changes, old behavior will
be maintained until `v1.6` (assuming no major releases).

When incompatible changes are made, old code paths will not get new
features. They simply become stale until removed. If you want to use a
new feature, and that section was involved in an incompatible change,
then you must move to new functionality to utilize new features.

As an example: the `email` configuration formatting changed in `v1.4`
(in the global configuration file). Additionally, `acceptInsecureCerts`
was added to `email` to support insure certificates. If you wish to use
this new setting, then you MUST update to the new `email` configuration.

As a result, it is recommended to immediately incorporate updates to
scripting or configuration files when upgrading to a new version.

** Note: This file only describes incompatible changes made in a release,
not ALL changes made in a release. Refer to release notes on GitHub for a
full description of all changes made, or new features that did not create
a compatibility issue. **

Finally, note that [README.md](README.md) only describes most current
configurations. If you're running the master branch before releases,
you should monitor issues raised and commits made so you know when changes
are made to the master branch.

### Changes made in v1.4

Incompatible changes in `v1.4` involve to areas of the code base:

1. Notifications (and how email is affected)
1. Command line changes

#### v1.4: Notifications

This release introduces notifications, an extendible mechanism to notify
of backup results. In the past, `duplicacy-util` only had the notion of
sending email. With extendable notificaitons, `duplicacy-util` can easily
have new forms of notifications added (push notifications to mobile devices,
desktop notificaitons, etc). This change involved removing email-specific
commnd line options (namely `-m` and `-tm`) and global configuration changes.

The `-m` (send mail) option was deprecated because, with notifications, that
behavior is automatic. You don't need to specify a command line option to
trigger notifications.

The `-tm` (test mail) option was deprecated because it only tests for email.
A new option, `-tn` (test notifications), was created to test notifications.
Note that `-tn` only works if you actually set up notifications in the
global configuration file.

Notifications are triggered at certain times in the backup process:

| Type      | When triggered |
| ----      | -------------- |
| onStart   | Triggered at the beginning of a backup |
| onSkip    | Triggered when a backup is skipped (if already running) |
| onSuccess | Triggered when a backup completes successfully |
| onfailure | Triggered when a backup fails due to an error |

If you wish to be notified, mention the notification type in a
comma-separated (if multiple notifications are to be triggered) list.
Be aware that, in `v1.4`, only `email` notifications are supported. If
you do not wish to be notified, do not specify the notification type
in the notifier.

Notifications are specified in the global configuration file with a
section like:

```
notifications:
  onStart: []
  onSkip: ['email']
  onSuccess: ['email']
  onFailure: ['email']
```

In this example, no notification would be triggered when a backup was
started, but email would be sent when a backup was skipped, succeeded,
or failed.

Additionally, the `email` configuration was changed as follows:

| Old email setting | New email setting |
| ----------------- | ----------------- |
| emailFromAddress | fromAddress (in section email) |
| emailToAddress | toAdddress (in section email) |
| emailserverHostname | serverHostname (in section email) |
| emailServerPort | serverPort (in section email) |
| emailAuthUsername | authUsername (in section email) |
| emailAuthPassword | authPassword (in section email) |

Thus, if you had an email configuration such as:

```
emailFromAddress: "Donald Duck <donald.xyzzy@gmail.com>"
emailToAddress: "Donald Duck <donald.xyzzy@gmail.com>"
emailServerHostname: smtp.gmail.com
emailServerPort: 465
emailAuthUsername: donald.xyzzy@gmail.com
emailAuthPassword: gaozqlwbztypagwt
```

then you should replace that with the following section:

```
email:
  fromAddress: "Donald Duck <donald.xyzzy@gmail.com>"
  toAddress: "Donald Duck <donald.xyzzy@gmail.com>"
  serverHostname: smtp.gmail.com
  serverPort: 465
  authUsername: donald.xyzzy@gmail.com
  authPassword: gaozqlwbztypagwt
```

Note that old `email*` flags will be removed in `v1.6`. Additionally,
new email options are available, but you must use new format to utilize
them.


#### v1.4: Command Line Changes

New command line options:

| Option | Purpose |
| ------ | ------- |
| -backup | Specifically perform a `duplicacy backup` operation. Can be combined with `-copy`, `-check`, and `-prune`. |
| -copy | Specifically perform a `duplicacy copy` operation. Can be combine with `-backup`, `-check`, and `-prune`. |
| -check | Specifically perform a `duplicacy check` operation. Can be combined with `-backup`, `-copy`, and `-prune`. |
| -prune | Specifically perform a `duplicacy prune` operation. Can be combined with `-backup`, `-copy`, and `-check`. |
| -tn | Test notifications (trigger each notification type configured in global configuration file).

Deprecated command line options:

| Option | Purpose |
| ------ | ------- |
| -b | Performed a `duplicacy backup` operation. Use `-backup` instead. |
| -c | Performed a `duplicacy check` operaiton. Use `-check` instead. |
| -p | Performed a `duplicacy prune` operation. Use `-prune` instead. |
| -tm | Triggered test email messages. Configure notifications in global configuration and use `-tn` instead. |

Note that deprecated command line options will be removed in `v1.6`.
