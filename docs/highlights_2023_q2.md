Highlights 2023 Q2 - etcd
======
<span style="color: #808080; font-family: Babas; font-size: 1em;">
ahrtr@github <br>
Auguest 8, 2023
</span>

# Table of Contents
- **[Background](#background)**
- **[Onboarded 7 new members and 1 previous maintainer back](#onboarded-7-new-members-and-1-previous-maintainer-back)**
- **[SIG-ifying etcd](#sig-ifying-etcd)**
- **[Completed tier-1 support for arm64](#completed-tier-1-support-for-arm64)**
- **[Published etcd roadmap](#published-etcd-roadmap)**

# Background
This post briefly summarizes the big changes in etcd community in 2023 Q2.

# Onboarded 7 new members and 1 previous maintainer back
The community onboarded 7 new members below,
- chaochn47
- jmhbnz
- fuweid
- tjungblu
- cenkalti
- pavelkalinnikov
- lavacat

`jmhbnz` was also promoted to a reviewer recently.

One of previous maintainers `wenjiaswe` is back, and was added as a maintainer in [pull/16197](https://github.com/etcd-io/etcd/pull/16197).

# SIG-ifying etcd
The PR [kubernetes/community/pull/7372](https://github.com/kubernetes/community/pull/7372) of setting up SIG-etcd is still
under review, but it's basically settled. There is no any change on the code base, but etcd will follow some Kubernetes's
processes or requirements, refer to the discussion the above PR and also this [Checklist for setting up SIG-etcd](https://docs.google.com/document/d/1JGpsDlQui6UcOnARk3Hvq-kYpWWmSQZTvs7FGXhYKdA/edit?pli=1&resourcekey=0-ip9ms08vN1JsOnZdPGJL_g).

# Completed tier-1 support for arm64
`Linux + arm64` is on tier 1 support, which means that it's now fully supported by etcd maintainers. Please read
[Supported platforms](https://etcd.io/docs/v3.5/op-guide/supported-platform/).

# Published etcd roadmap
We added a [roadmap](https://github.com/etcd-io/etcd/blob/main/Documentation/contributor-guide/roadmap.md) for etcd
recently. The etcd community will continue to mainly focus on technical debt over the next few major or minor releases.
