I spoke to Chris Harden, Justin Lai, Paul Carlton, Steve Fraser, Chris Lavery
and Darryl Weaver (a group of people who are using WeGo in anger) about their
experiences with Weave GitOps documentation.

The aim of these conversations was to understand how customers/users experience
our documentation and discover which bits are helpful, which were not, and which
were just plain absent.

The installation of EE (as well as the upgrade path from OSS->EE) was deliberately
deprioritised in 2022 in favour of getting demoable features in front of potential
customers. Now we have customers. Anecdotal evidence suggests that WGEE is not
intuitive or easy to get up and running. This will not be entirely the fault of
the docs, nor will it be entirely solved by improved docs, but there will
definitely be opportunities to make the process clearer via documentation while
we engineer an easier path in 2023.

My starting point questions sought to find out the following: was there any part
of the documentation which 'just worked', which we can use as a baseline for
'good docs'; which parts were difficult to follow; what are we missing.

**Note that conversations were centered on EE documentation, as that is what our
CX/CRE teams are spinning up.** Also >80% of the docs are just EE anyway.

The marker of perfect docs is that they are taken for granted. If they work
well, you barely notice them. People tend to only talk about docs when they do
not deliver. Currently this is where we are at.

I'd like to note here that **our docs are not bad**. Engineering has done an
admirable job of adding content where and when they could. The issues we feel
now are because if documentation is not done intentionally and seen as a product
feature itself, it becomes an addon and an afterthought. Content which is
quickly added by developers after delivering a feature will never be as good as
a deliberately crafted narrative of the product. BUT if you establish a good
baseline in which all the key structures and stories are there, then it becomes
safe and easier for engineers to throw stuff in as they think of it.

## What are our people using the docs for?

Few people read docs cover to cover. After following initial getting started
tutorial or guides, most will swing back to look up specific feature or API
information. My interviewees tended to come for general reference or new feature
details.

## What are we doing right?

Unfortunately this triggered a lot of silent head scratching. This is not to say
the docs are not usable, but we don't have any examples where everything works
perfectly with clear explanations and fool-proof copy-pastable examples.

The OSS 'Getting Started' guide was the only area which got a thumbs-up, but it
was noted that the EE guide was much less intuitive.

## What were the problems raised in the interviews?

I am going to split these up into 'exists but confusing', 'completely missing
information', and 'general gripes/concerns'.

### Content which exists but is imperfect

- Policy documentation is very hard to follow.
- The relationship between GitOpsSets and pipelines is not clear.
- Tenancy (specifically RBAC in leaf and management clusters) is a nightmare to
  understand.
- There are some areas in which there are no examples of how to set things up,
  or there are fragmentary examples.
- Several copy-pasta commands have typos and therefore fail.

### Required content that we are missing

- EE Helm Chart reference
- Quickstart guides for EE. For example one for getting up and running with
  CAPI, one for the tf-controller. With working examples, and/or fully cloneable
  repos.
- Likewise, quickstarts for cluster management. One guide for setting up with an
  existing cluster, simply managing apps and pipelines. Another for cluster
  provisioning.
- Guides or cloneable examples for a 'correct' or recommended folder structure.
  There is little guidance on what a 'good layout' looks like and a baseline
  would be good.
- A proper guides section, written/titled around outcomes not
  implementation/tools.
- A 'narrative' or 'story' of how all the features tie together.
- API references for all WGEE/OSS objects.
- Examples for setting up OIDC with various providers.
- Perhaps even some terraform for setting up OIDC integrations.
- CLI based docs. We are GUI heavy right now. The GUI is good for selling to
  managers, but the people using the product will want to use the CLI.
  (Conversely, there are some sections which are only CLI based. So more balance
  overall would be great.)
- Links from our cluster management section to relevant areas of official CAPI
  docs would be useful to fill in information gaps.

### General comments

- We need more automation in the docs. There are still so many manual steps, we
  should provide some terraform for setting up a cluster with WGEE, with some
  cluster secrets, with some OIDC. Basically for getting started guides (all
  giudes) we need as few manual steps as possible.
- Doubtful that anyone will be able to get WGEE set up within half a day, or
  even a day, following this documentation.
- Concerns that the tf-controller section of docs are falling out of step with
  the main tf-controller docs.
- The VSCODE extension is only mentioned once.
- Pipelines do not seem to add any value over the Flux-only way, in fact there
  is an extra step. Templates which provide a full setup, and docs which make
  this look easy, would fix this.
- Nearly all the sections assume a lot of knowledge from our users.
- Guides on how to integrate with other things are more important than getting
  started steps.
- The documentation does not make the product easy for App developers, we focus
  heavily on the Ops experience.
- The versioning is deeply confusing, since most of the docs are EE but the
  versions are for OSS.

### Product comments

These notes are more about product features than they are about the docs. I am
noting them here, but they are not relevant for the rest of this page.

- We don't release based on actual semver, meaning that breaking changes occur
  on minor releases. Until we follow semver, we cannot get customers to
  auto-bump versions.
- There should not be a separate EE CLI tool.
- Whatever you can do in the CLI you should be able to do in the GUI. Right now
  the GUI does not give me the full experience.
- We force generic command-line users to need 4+ different CLI tools (gitops,
  kubectl, flux, clusterctl), export variables, run all these commands, etc. Why is there
  not just something like a toml file? Users fill that in, run one command, and
  we do the rest.
- We should not be enabling telemetry by default.

## How can we improve things in the short term?

A lot of the short term mitigations are what we have recently done/are doing:

- Checking that tutorials/examples/commands are all correct and copy-pastable
- Improving and clarifying language
- Improving layout and making things easier to find
- Rearranging content to make it accessible and easy to follow
- Re-taking screenshots so that all information is current
- Making sure all commands and examples can be copy-pasted without issue

These are all good things to have done to hold us over while we do the heavy
lifting to solve the problems which can't be fixed my moving existing content
around.

## What are the next steps?

A few of the issues raised can be quickly solved by opening tickets:

- Ask CX/CRE to review Policy docs after recent review (follow up tickets if
  work still needed)
- Develop a 'baseline' recommended folder structure, with an example repo to
  clone
- Add links in all CAPI cluster docs to the relevant official CAPI docs
- Clarify the relationship between GitOps sets and Pipelines
- Create an example repo showing a tenancy setup between a management and a leaf
  cluster including full RBAC
- Add API references of GitOps objects to the docs
- Add examples for how to set up OIDC with commonly used providers
- Create EE quickstart tutorials
- Create a Pipeline template which delivers a single 'click' setup
- Open source the EE helm chart and add a reference to the docs. Better yet
  don't have a separate chart at all, just have some way of enabling/disabling
  EE features by entitlement.

The rest are larger and will only be made better by a) rethinking how we write
documentation, b) thinking about who we are writing it for, and c) putting
processes in place to ensure quality is maintained into the future.

### General documentation structuring

[Here is a handy guide](https://documentation.divio.com/) which lays out a good model
and set of principles for docs writing.

### Versioning

This has been discussed a lot, and we have settled on keeping all versions up to
the earliest one known to be used by a customer, and revisiting later (like a
year from now).

For context, reasons why versioning is a pain:
- We technically have 3 different products in our docs (OSS, EE, tf-controller)
  and only one of those lines up with the versions selectable in the docs (more
  on this story later).
- `Edit this page` links always take you to current.
- Contributing in general is a pain so when we build a community we have an
  obstacle.
- Bugs like this can happen: https://weaveworks.slack.com/archives/C04KNEME4AJ/p1676437989871869
- All these historical versions mean the site takes aaaaggees to build.

Reasons why we want versioning:
- We have a very new product and need to keep customers happy. All the info for
  all the versions needs to be there.

### OSS vs EE

This comes up a lot and the gist of complaints are:
- The docs are versioned to OSS
- But the majority of them refer to EE features
- It is not intuitive at all to know which version of the docs relates to which
  version of EE
- (Another thing to note is that the tf-controller docs are also here... and
  also versioned at a completely different cadence from the docs.)

Basically if you are an EE customer and you think 'Oh I am on version 0.15.0,
I'll read the docs at that version', the docs you end up reading were versioned
at OSS 0.15.0 and _whatever the EE version happened to be at that time_. So you
are reading something completely unrelated. As CX pointed out: at this stage we
need to make sure that every EE customer can easily find the docs for their
version, and that is just not possible right now.

**We have argued that versioning is important and something we want to continue
doing, but it this current state it is worse than useless.**

A possible solution to this has been suggested: that we release EE on the same
schedule as OSS. This may be a good idea if EE was the same product/project as
a plugin or extension of OSS, but it isn't. It is basically a whole other
product which uses OSS libs: you don't even need to install OSS first.

We have so far been modelling our docs on Gitlab's: all documentation in one
place, with Tier labels marking the difference. The key difference is that
Gitlab [**has a single public codebase containing both OSS and EE offerings**](https://about.gitlab.com/blog/2019/08/23/a-single-codebase-for-gitlab-community-and-enterprise-edition/).
(It may be worth reading [exactly how they did that](https://gitlab.com/gitlab-com/gl-infra/readiness/-/blob/master/library/merge-ce-ee-codebases/index.md)
in case we need some inspo.)

Because we don't hold code like Gitlab, trying to write docs like Gitlab is just
not going to work the same for us.

So what do we do here? Some options in no particular order.

1. Copy gitlab, try to find a way to have one codebase.
    - Their exact method will not work for us
1. Have EE on the same release schedule as OSS
    - Problematic when EE needs to suddenly release some security patches
1. Don't version the docs at all, given they have no intuitive relevance to the
   EE version
    - This come up a lot and we have decided that there is value in keeping
      versions, but versioning is pointless if they do not relate to the product
      version you are on.
1. Don't change anything, but make it really clear that the versions do not
   directly reference EE versions, and update Release notes page to state which
   docs version they correspond to.
    - This will be so easy to miss and I can see customers thinking "oh I am on EE
      version `0.15.0`, I'll read the docs version `0.15.0`.1
1. Have the docs version follow the EE version (since that is the majority of
   the docs anyway).
1. Put all Enterprise docs in a separate sidebar, linked from the top nav.
    - Does not solve any versioning confusion, BUT does give us a clear line
      between docs to highlight for users.
1. Have a docs instance for EE. With Docusaurus you can have multiple areas or
   "instances" of your docs which can be versioned **separately**. I have
   spiked this out (somewhat imperfectly) and left information on [the spike](https://github.com/weaveworks/weave-gitops/pull/3495).

### Terraform Controller

This is an extremely valuable component which has a lot of heat, certainly more
than WGEE.

The tf-controller's docs are in the weave-gitops docs. The tf-controller
versions have nothing to do with the documentation versions. So again we are in
a position where a customer is using tf-controller 0.10.0, so they go to the
docs version 0.10.0... and those docs are completely irrelevant.

The choices here are similar to the EE vs OSS decision above:

1. Keep things as they are, add banners etc to make it clear which tf-controller
   version is documented where.
1. Move tf-controller docs out to their own dedicated site.
1. Create a separately versioned instance in the existing site.

### Platform Operator vs App Developer

We have distinct personas who used our product.

So far the majority of our documentation caters to Platform Operator types (ie
the people who will be deploying WGEE, setting up tenancy, adding profiles, etc).
There is some information there for Application Developers (ie the people given
a tenancy, access to profiles etc), but it is somewhat
buried and tricky to find information that you would need in that capacity.

The idea of writing 'persona orientated' docs has been floated. It was noted in
one of the interviews that they "have never seen persona-based docs, but
wouldn't that be sweet?".

I don't propose doing persona-based docs super granularly, but structuring
around 2 activities (setting up WGEE for people in your org, using WGEE as a
tenant) could give a very smooth experience for user onboarding. Developers who
have been given access to a platform by their Ops team can quickly find the
information they need without having to scroll past all the setup guides;
Operators will find a clear boundary between 'set the thing up' and 'okay now
check that it works'.

Some ways we could do this:
- Rearrange the sidebar into 2 groups 'For Operators', and 'For Developers' and
  shuffle information under there accordingly.
- Update what we already have with separated clear pages/subsets for setting up and then
  using.

### GUI vs CLI

It has been mentioned that our docs are quite GUI heavy (or they have only one
option rather than both). It is very likely that the majority of users will either
go straight for a CLI option, or will end up there anyway.

We need to ensure that for every example or walkthrough, both the CLI and the
GUI methods are covered. I don't think any structural changes need to happen
here, but it needs to factor into our process.

This is especially important because some customers have not even looked at the
GUI yet because all they want is CLI.

There is a question here about implementation: the most obvious way to do this
is to use 'tabs' showing a CLI option and a GUI option. But these tabs are not
linkable, so it becomes harder to share a URL with customers.

### Guides and best practices

Useful guides take users through several easy-to-follow and adaptable steps
towards a clear outcome. They will cover at least one product feature, or
combine many, and could include integration with some external component.
Guides are often known as 'advanced' because they are taking users through a
near real life use-case, far beyond 'simply set up this one thing'.

We have some guides right now, some are [thorough and have clear goals](https://docs.gitops.weave.works/docs/next/guides/delivery/), some
seem to suddenly [stop half way through](https://docs.gitops.weave.works/docs/next/guides/deploying-capa/)
with barely a mention of WGEE, some assume quite a bit of knowledge from the
user with no links for them to go get that knowledge, and some advise on
[standard kubernetes things which are not (explicitly) related to WGEE](https://docs.gitops.weave.works/docs/next/guides/cert-manager/).
All of them are somewhat lightweight, and don't provide users with much
opportunity to really sink their teeth into the product, which **really is what
they should be doing**. Many of these guides could easily be moved into general
setup or configuration sections.

The purpose of these guides should be to explain things that users would not be
able to find anywhere else: in other words specific use-cases for our product.

Examples of guides that we could do with that we are now missing:
- How to build a recommended directory structure for:
  - single cluster with some apps
  - management+leaf cluster with some apps
  - management+leaf cluster with semi-complex tenancy
  - (This should actually be a whole subsection under the guides)
- How to configure RBAC, OIDC and tenancy for a management+leaf cluster.
- How to go from having zero policies in place to several policies across more
  than one cluster and tenancy.
- How to use GitOpsSets and Pipelines to manage promotion of multiple apps
  across multiple environments.

**All of the above**, and ideally all guides in general, should have a repo or template
which can be forked/cloned and 'launched' locally within an hour. We should also
take care to title based on outcome, not implementation or technology, eg: 'How
to set up SSO' over 'Configure OIDC'.

A further step would then be to combine several of these tips into larger "best
practices" guides, eg around GitOps at scale.

In the future it would be good to get feedback from CX/CREs to find out what
they are trying to do that the docs are NOT currently helping them with, and
working with them to produce that advanced content. Which brings me on to...

### Cross-department writing

Engineers are using the product in completely different ways from our users and
from CX/CRE. Thus they are often writing the documentation from the position of: I
have finished developing this one subset of feature X, I have bootstrapped a
local environment to test and write about feature X's happy-path, I am not
simultaneously testing or writing about any other feature in conjunction with
this feature, and I am doing this all on a local environment.

Customers and CREs are often using several features in combination, which is
exactly what we want them to be doing, which means that the information in the
docs doesn't quite line up with what they are doing.

This is why I think engineering could do with some assistance from people who
have actually used all/many facets of the product at the same time to actually
write useful guides. This could be in the form of writing content and having
someone from CX/DX/CRE review; pairing; or direct contributions from those who
have already muddled through and come up with a solution.

Engineers are busy churning out features, CREs are busy getting customers
onboarded, nobody has time. But if we want anyone to have time to do literally
anything else in the future, many departments are going to have to work together
to get (and keep) the docs to a level where the customers **do not need to ask
for help**.

And so we get on to...

### The long term process

If everything I have detailed above is put into place, I feel that we have a
pretty maintainable baseline. But it takes work to keep quality high.

Moving forward, I suggest that the docs be treated as a unique **feature of weave
gitops**, because that is what they are. This means not throwing a 'oh and also
do docs' bullet or ticket for some tired engineer to do later, or as solely a
part of another feature's work.

So that means:
- initiatives, epics, or tickets with concrete goals
- gathering user (or CRE) feedback at least every quarter
- encouraging user engagement, from reporting issues to contributing
- involving more than just the engineer who developed the feature writing the
  content (it should be a team thing, with at least one person running through
  any steps or walkthroughs to actually use the thing)

Not everyone is comfortable with writing docs, which is fine, but we don't want
the responsibility falling on just the one or two people who are. In this case
pairing is fine, with one person providing the code/yaml examples or tutorial
and maybe some bullets, and another person filling in the language (and a
further person reviewing).

This all seems work-intensive, because it is. But if the effort is not put in
and the process not established, then our users will be noticing the docs for
all the wrong reasons.

(If I have time before I go, I will create some templates for how to write epics
or tickets for docs. I may not have time, so [this guide](https://documentation.divio.com) may be useful.)

## How can people report problems with the docs or request more content?

Well, from March 17th you can no longer personally message me about it ;)

The simplest way to get something fixed or added to the docs is by opening an
issue in the OSS weave-gitops repo using the provided documentation template.
Being thorough with the information you provide is key here, as is tagging in
someone from engineering who can make sure the change goes through.

For those who want to immediately make a change, we do have an `Edit this page`
button. This button will open the page at **current version** for editing. This
may sound not-ideal, and it could be argued that we put in a couple extra lines
of config in `docusaurus.config.js` to send the clicker to the actual version
they want to edit, but that means we get the fix in the older version and it is
not immediately carried over to the current or newer ones. Conversely, that
versioned page might no longer exist in the present. I would rather people
opened tickets pointing out the flaw, and then they or the team can contribute
to make the change across _all_ affected versions. This is just something we
have to grapple with when we version our docs.

In order for people to contribute we need to have very good 'contribute to the
docs' docs. The readme has been updated recently, but more work needs to be done
here, and around contributing in general.
