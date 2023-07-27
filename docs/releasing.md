# Releasing `weave-gitops-enterprise`

[comment]: <> (Github can generate TOCs now see https://github.blog/changelog/2021-04-13-table-of-contents-support-in-markdown-files/)

How to release a new version of weave-gitops-enterprise

## Versioning

We follow [semantic versioning](https://semver.org/) where
- Releases adding new features or changing existing ones increase the minor versions (0.11.0, 0.12.0, etc)
- Releases exclusively fixing bugs increase the patch version (0.11.1, 0.11.2)

## Prerequisites

Install [GnuPG](https://gnupg.org/) and [generate a GPG key and add it to your Github account](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key).

If you aren't using the GPG suite, you will need to [add the GPG key to your .bashrc / .zshrc](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key).

## Checklist

### Radiate that a release is happening (**30 mins** waiting for objections, but a day or two is better)

Write a message in `#weave-gitops-dev` on slack. Ideally we want to fall into a predicatable release cadence so that the release doesn't come as a surprise to anyone.

Wait and see if any team has an objection like a known release blocking bug.

Sample:

> Hi! We would like to to release WGE v0.9.6, does anyone have any concerns or reasons to delay the release?

### Make sure weave-gitops-enterprise on `main` is green (**30s** to check)

Look for the green tick next to the last commit on [weave-gitops-enterprise](https://github.com/weaveworks/weave-gitops-enterprise)

### Make sure all new OSS commands are included

There is a bot that automatically creates a PR to update the version of OSS used by WGE on every stable release of OSS. However, if there are any new commands introduced in OSS, these need to be added manually by an engineer. Therefore we need to make sure that the new OSS commands are available from WGE CLI too.

### Make sure all new features in the latest weave-gitops core release work

In the past we have released WGE without making sure the new features in Weave GitOps core have been integrated.

Review the release notes of the [latest version of core](https://github.com/weaveworks/weave-gitops/releases) and check the new features listed there work in WGE. If its confusing ask in `#weave-gitops-dev` on slack.

## Release
### Create a tag

Tag the new release with an **annotated** and **signed** tag.

```bash
cd weave-gitops-enterprise
git checkout main
git pull

# Make sure your local commit is the same as the head on github
git log

# Tag a push an annotated tag
git tag -a -s v0.9.4 -m "Weave GitOps Enterprise v0.9.4"
git push origin v0.9.4
```

This will kick off the release process in GitHub actions
- View progress via the [release action](https://github.com/weaveworks/weave-gitops-enterprise/actions/workflows/release.yaml).
- The action will generate release notes and publish a release on GitHub
- A message will be sent to #weave-gitops-dev in slack announcing the release

### Update release notes with *Dependencies* and *Highlights* and ensure the user guide references the new version

#### Github releases page

If this is a final (non-rc) release, de-select the `This is a pre-release` checkbox.

> **Note**
> If there are PRs listed here that you are not familiar with, ask in `#weave-gitops-dev` on slack.
> You can ping another team to ask them to update with highlights from their perspective.

You can use a [previous release](https://github.com/weaveworks/weave-gitops-enterprise/releases) as a template here.

- Edit the new release, copy and paste the **Dependency Versions** section from an older releases. Update any versions that have been changed.
- Add a **Highlights** section, calling out the significant changes in this new release.
- Add a **Breaking Changes** section, calling out any breaking changes in this new release.
- Add a **Known Issues** section, calling out any known issues in this new release.

If the previous release was an `rc.x` release, add notes from its release notes.

Notify your Product Manager at this stage that the release notes are available. They will combine with content from the product newsletter to update the website **after** the release.

#### The https://docs.gitops.weave.works/ Enterprise releases page

> **Note**
> This is for non-rc.x releases only

Copy the **Dependency Versions**, **Highlights**, **Breaking Changes** and **Known Issues** sections from the Github release notes into the [Enterprise releases page](https://github.com/weaveworks/weave-gitops/blob/main/website/docs/enterprise/getting-started/releases-enterprise.mdx).

- Paste it up the top, keeping previous releases in order.
- Add the current date under the release version

Make sure to backport the docs changes to the most recent versioned docs. For example when releasing 0.9.5, both these files should be the same:
- weave-gitops/website/docs/enterprise/getting-started/releases-enterprise.mdx
- weave-gitops/website/versioned_docs/version-0.9.5/enterprise/getting-started/releases-enterprise.mdx

#### Make sure the installation docs are updated for the upcoming version

The following pages reference the version:
- https://docs.gitops.weave.works/docs/installation/weave-gitops-enterprise/#install-cli
- https://docs.gitops.weave.works/docs/installation/weave-gitops-enterprise/#5-configure-and-commit

Ensure that the version referenced in the instructions for downloading the CLI and the version used in the WGE Helm release example, match the upcoming version.

### Announce final (non-rc) releases in #weave-gitops on Weaveworks slack

Sample:

> Hi! There is a new release of Weave Gitops Enterprise v0.9.6!
> - https://github.com/weaveworks/weave-gitops-enterprise/releases/tag/v0.9.6
> - https://docs.gitops.weave.works/docs/enterprise/getting-started/releases-enterprise/


### Update this document with any thing that unclear!

Always be improving this document. If you find something that is unclear, or something that could be improved, please update this document and send a PR.
