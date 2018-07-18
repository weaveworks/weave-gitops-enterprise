# Weave Kubernetes Subscription (WKS)

![Service description diagram](https://www.weave.works/assets/images/blt1670b4d9d8010619/KB_support_diagram.jpg)

## Purpose
This repository is to keep together work done on the Weaveworks Kubernetes Subscription. Track the progress in the [Github project](https://github.com/weaveworks/wks/projects/1).

## important documents
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

## Notes

**`tools/`**

The `tools` directory is copied via `git subtree` from the
[build-tools](https://github.com/weaveworks/build-tools) repo.

**code-generator**

```
$ make gen

```
