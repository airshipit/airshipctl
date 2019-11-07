# Contributing Guidelines

The airshipctl project accepts contributions via gerrit reviews.  For help
getting started with gerrit, see the official [OpenDev
documentation](https://docs.openstack.org/contributors/common/setup-gerrit.html).
This document outlines the process to help get your contribution accepted.

## Support Channels

Whether you are a user or contributor, official support channels are available
[here](https://wiki.openstack.org/wiki/Airship#Get_in_Touch)

You can also report [bugs](https://airship.atlassian.net/issues/?jql=project%20%3D%20AIR%20AND%20issuetype%20%3D%20Bug%20order%20by%20created%20DESC).


Before opening a new issue or submitting a patchset, it's helpful to search the
bug reports above - it's likely that another user has already reported the issue you're
facing, or it's a known issue that we're already aware of. It is also worth
asking on the IRC channels.

## Story Lifecycle

The airshipctl project uses Jira to track all efforts, whether those are
contributions to this repository or other community projects. The Jira issues
are a combination of epics, issues, subtasks, bugs, and milestones.  We use
epics to define large deliverables and many epics have been created already.
The project assumes that developers trying to break down epics into managable
work will create their own issues/stories and any related subtasks to further
breakdown their work. Milestones act as human readable goals for the sprint they
are assigned to.

- [Active Sprints](https://airship.atlassian.net/secure/RapidBoard.jspa?rapidView=1)
- [Issues](https://airship.atlassian.net/projects/AIR/issues)

The airshipctl project leverages 1-month sprints primarily for the purpose of
chronologically ordering work in Jira.

### Coding Conventions

Airship has a set of [coding conventions](https://airship-docs.readthedocs.io/en/latest/conventions.html) that are meant to cover all Airship subprojects.

However, airshipctl also has its own specific coding conventions and standards in the official airshipctl [developer guide](docs/developers.md).