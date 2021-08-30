# Releasing `weave-gitops-enterprise`

[comment]: <> (Github can generate TOCs now see https://github.blog/changelog/2021-04-13-table-of-contents-support-in-markdown-files/)

How to release a new version of weave-gitops-enterprise

## Create a tag

_This may be possible by creating a tag via the releases system in the github UI, please test and update here._

Tag the new release with an **annotated** and **signed** tag.

```bash
git tag -a -s v0.0.6 -m "Weave GitOps Enterprise v0.0.6"
git push origin v0.0.6
```

CircleCI will build the new release based on the new tag.

## Create a release in GitHub

- Go to the **Releases page** for the weave-gitops repository
- Click on **Draft a New Release**
- Add the tag you just pushed
- Add some release notes. It would be nice to automate this in the future. Its probably easiest to copy the last release as a template and then modify the list of _fixes_ and _new features_.
- Click on **Publish Release**
