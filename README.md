# Weave Kubernetes Subscription (WKS)

![Service description diagram](https://www.weave.works/assets/images/blt1670b4d9d8010619/KB_support_diagram.jpg)

## Purpose
This repository is to keep together work done on the Weaveworks Kubernetes Subscription. Track the progress in the [Github project](https://github.com/weaveworks/wks/projects/1).

## Important documents
See:
- [Meeting Notes](https://drive.google.com/open?id=1wfN4V6T9t1-eapXGabFZqkBCxyKW3uVZzz-cBCosgxs)
- [Phase 1 Plan](https://docs.google.com/document/d/1q3y0jDrzNKpTxPUi5JYf8vaPDTLV9_Ur65lxZFElDSo/edit)
  - [Pharos/WKS analysis](https://docs.google.com/document/d/1FRJd5Uj0CuHPwHbqXooIpUF1UKTy9tjsBaNqAA5BtrQ/edit)
  - [Test plan](https://docs.google.com/spreadsheets/d/1EdSdbdbFrYrjLwr33qAMF31n_g2hrSgogljen8RBHj4/edit)
- [WKS manifest draft](https://docs.google.com/document/d/1WtIE11RC-6f4mhp2Krsf1AsNCNEHcSuEQNp12nV0mDU/edit#)
- [Press release](https://www.weave.works/press/releases/weaveworks-launches-enterprise-gitops-services/)
- [Product Page](https://www.weave.works/product/enterprise-kubernetes-support/)
- [Theory of Gitops](https://docs.google.com/document/d/1Y8kr3gROHUnFuGR3h4adjwWH6E3ttGHIYwVuWWVv2VE/edit)
- [WKS Future](https://docs.google.com/document/d/1HK6r5CA0ZlUQT3PmFWVQ_93TlPz31nHdx13-pve1S4U/edit#)

## Kerberos

- [Explain like I'm 5: Kerberos](http://www.roguelynn.com/words/explain-like-im-5-kerberos/)
- [About Kerberos Principals and Keys](https://ssimo.org/blog/id_016.html)
- [MIT Kerberos Documentation](http://web.mit.edu/kerberos/krb5-1.12/doc/index.html)

## Notes

**Releasing**

To release a new version of the project:
  - Create a new tag: `git tag -a 1.0.1`
  - Push tag: `git tag --push`
  - CI will push binary to weaveworks-wks.s3.amazonaws.com/wksctl-1.0.1
  - Edit release notes https://github.com/weaveworks/wks/releases/edit/1.0.1
  - Update rpm/wksctl.spec version and changelog
  - Build an rpm `cd rpm && ./build wksctl.spec`
  - Sign rpm: `rpm --addsign output/x86_64/wksctl-1.1.0-0.x86_64.rpm`
  - Publish rpm to our yum repo https://github.com/weaveworks/rpm
    - Copy rpm in `wks/rhel/7`
    - `cd wks/rhel/7 && createrepo .`

**`tools/`**

The `tools` directory is copied via `git subtree` from the
[build-tools](https://github.com/weaveworks/build-tools) repo.

**code-generator**

```
$ make gen

```
