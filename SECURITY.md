# Security policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately through GitHub's **Report a vulnerability** feature on this repository:

<https://github.com/MIna-Maher/k8s-diff-informer/security/advisories/new>

Do not open a public issue for an unpatched vulnerability. If GitHub does not let you submit a private report, contact the maintainer through their [GitHub profile](https://github.com/MIna-Maher) to arrange a secure channel. Do not post vulnerability details publicly or include credentials or unrelated personal data in the report.

Please include the affected version or commit, the impact, steps to reproduce, and any suggested mitigation. The maintainer will acknowledge the report and coordinate a fix and disclosure timeline with you.

## Supported versions

Security fixes are currently made on the default branch. There is no stable release yet; published beta images are not guaranteed to receive backported fixes. Use the latest source and review release notes before upgrading.

## Sensitive data

The informer reads Kubernetes resources and sends selected diffs to Slack. Resource values may be sensitive. Limit watched resources and access, configure ignored fields, and protect Slack destinations. Never publish real webhook URLs, kubeconfigs, tokens, or private cluster data in issues, logs, or pull requests.
