# Contributing Guidelines

The airshipctl project accepts contributions via Gerrit reviews.  For help
getting started with Gerrit, see the official [OpenDev
documentation](https://docs.openstack.org/contributors/common/setup-gerrit.html).
This document outlines the process to help get your contribution accepted.

## Support Channels

Whether you are a user or contributor, official support channels are available
[here](https://wiki.openstack.org/wiki/Airship#Get_in_Touch).

You can also request features or report bugs
[here](https://github.com/airshipit/airshipctl/issues/new/choose).

Before opening a new issue or submitting a change, it's helpful to search the
bug reports above - it's likely that another user has already reported the
issue you're facing, or it's a known issue that we're already aware of. It is
also worth asking on the IRC channels.

## Story Lifecycle

The airshipctl project uses
[GitHub Issues](https://github.com/airshipit/airshipctl/issues) to track all
efforts, whether those are contributions to this repository or other
community projects. The GitHub Issues are a combination of epics, issues,
bugs, and milestones.  We use epics to define large deliverables that need to
be broken down into more manageable chunks. Milestones act as human readable
goals for the sprint they are assigned to.

### Coding Conventions

Airship has a set of [coding conventions](
https://docs.airshipit.org/develop/conventions.html) that are meant
to cover all Airship subprojects.

However, airshipctl also has its own specific coding conventions and standards
in the official airshipctl [developer guide](
https://doc.airshipit.org/airshipctl/developers.html).

### Submitting Changes

All changes to airshipctl should be submitted to OpenDev's Gerrit. Do not try
to fork the repository on GitHub to submit changes to the code base.

All issues are tracked via
[GitHub Issues](https://github.com/airshipit/airshipctl/issues) and are tagged
with a variety of helpful labels. If you are new to the project, we suggest
starting with issues tagged with
"[good first issue](https://github.com/airshipit/airshipctl/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)"
to help get familiar with the codebase and best practices for the project.

When you find an issue you would like to work on, please comment on the issue
that you would like to have it assigned to you. A project admin will then
make sure that it is not currently being addressed and will assign it to you.

As you work on an issue, please be sure to update the labels on it as you work.
When you start work on an issue, use the "wip" label to indicate that you have
begun on a change for the issue. When your work is completed, submit a comment
to the issue with a link to your change on Gerrit and change the "wip" label
to a "ready for review" label to indicate to the community that you are
seeking reviews.

In your commit message, be sure to include a reference for the issue you
are addressing from GitHub Issues. There are three ways of doing this:

1. Add a statement in your commit message in the format of `Relates-To: #X`.
This will add a link on issue "#X" to your change.
2. Add a statement in your commit message in the format of `Closes: #X`.
This will add a link on issue "#X" to your change and will close the issue when
your change merges.
3. Add a bracketed tag at the beginning of your commit message in the format of
`[#X] <begin commit message>`. This will add a link on issue "#X" to your
change. This method is considered a fallback in lieu of the other two methods.

Any issue references should be evaluated within 15 minutes of being uploaded.

**NOTE** Make sure to carefully divide the work into logical chunks to avoid
creating changes that are too large. Such practices are discouraged and make
code review very difficult. Break down the work into components and create a
separate change for each component. Keep a design document or README to
track the overall progress when making a large contribution.
See [OSF Guidelines](https://docs.openstack.org/contributors/code-and-documentation/patch-best-practices.html#the-right-size) for more information.

## Reviewing Changes

Another great way to contribute to the project is to review changes made by
others in the community. To find changes to review, you can filter by ready
for review on GitHub Issues or you can search Gerrit for open changes.
Links to both of these can be found below:

* [GitHub Issues "ready for review" Filter](https://github.com/airshipit/airshipctl/issues?q=label%3A%22ready+for+review%22)
* [Gerrit Review Board for airshipctl](https://review.opendev.org/#/q/status:open+NOT+label:Verified%253D-1+NOT+label:Workflow%253D-1+NOT+message:DNM+NOT+message:WIP+project:airship/airshipctl)
