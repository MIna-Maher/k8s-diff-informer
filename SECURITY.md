# Security policy

## Reporting a vulnerability

Please report suspected vulnerabilities privately through GitHub's **Report a vulnerability** feature on this repository:

<https://github.com/MIna-Maher/k8s-diff-informer/security/advisories/new>

Do not open a public issue for an unpatched vulnerability. If GitHub does not let you submit a private report, contact the maintainer through their [GitHub profile](https://github.com/MIna-Maher) to arrange a secure channel. Do not post vulnerability details publicly or include credentials or unrelated personal data in the report.

Please include the affected version or commit, the impact, steps to reproduce, and any suggested mitigation. The maintainer will acknowledge the report and coordinate a fix and disclosure timeline with you.

## Supported versions

Security fixes are made on the default branch and included in future releases as appropriate. Previously published beta images are not guaranteed to receive backported fixes. Use the latest stable release and review its notes before upgrading.

## Sensitive data

The informer sends selected resource update diffs to the configured Slack destination. Secret resource `data` and `stringData` values are always redacted. Common credential-like field names and configured redaction paths are masked while preserving an indication that the field changed. Configure `RESOURCES_REDACTED_FIELDS` (or Helm's `config.fieldsToRedact`) for application-specific paths; `*` matches one path segment, including a list index. Escape a literal dot in a map key with `\.`.

Credential-name detection is best-effort and cannot recognize every secret embedded in arbitrary strings or custom resource fields. Limit watched resources, use explicit redaction paths for sensitive configuration, and restrict access to Slack destinations. Never publish real webhook URLs, kubeconfigs, tokens, or private cluster data in issues, logs, or pull requests.
