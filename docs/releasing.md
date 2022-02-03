# Releasing `weave-gitops-enterprise`

[comment]: <> (Github can generate TOCs now see https://github.blog/changelog/2021-04-13-table-of-contents-support-in-markdown-files/)

How to release a new version of weave-gitops-enterprise

## Prerequisites

Install [GnuPG](https://gnupg.org/) and [generate a GPG key and add it to your Github account](https://docs.github.com/en/authentication/managing-commit-signature-verification/generating-a-new-gpg-key).

If you aren't using the GPG suite, you will need to [add the GPG key to your .bashrc / .zshrc](https://docs.github.com/en/authentication/managing-commit-signature-verification/telling-git-about-your-signing-key).

## Create a tag

_This may be possible by creating a tag via the releases system in the github UI, please test and update here._

Tag the new release with an **annotated** and **signed** tag.

```bash
git tag -a -s v0.0.6 -m "Weave GitOps Enterprise v0.0.6"
git push origin v0.0.6
```

CircleCI will build the new release based on the new tag.
