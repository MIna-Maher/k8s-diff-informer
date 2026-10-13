---
layout: default
title: k8s-diff-informer
description: Know what changed in selected Kubernetes resources with filtered, readable Slack resource change diffs.
---

# k8s-diff-informer

<p><img src="{{ '/assets/logo.svg' | relative_url }}" alt="k8s-diff-informer logo" width="96"></p>

**Know what changed, without digging through Kubernetes events or comparing YAML by hand.** Choose the resources to watch, ignore noisy fields, and get Slack resource change notifications with readable field diffs for additions, meaningful updates, and deletions. Built for platform and Kubernetes operators who want a lightweight change feed for selected resources.

The tool reports object-level resource changes. Use your monitoring and alerting stack to evaluate cluster health and conditions. The Go service includes retrying Slack delivery, Prometheus metrics, and health endpoints.

![Architecture and event flow]({{ '/assets/architecture.svg' | relative_url }})

## Get started

- [Install with Helm]({{ '/install.html' | relative_url }})
- [Configure monitoring]({{ '/monitoring.html' | relative_url }})
- [Read the repository README](https://github.com/MIna-Maher/k8s-diff-informer#readme)
- [View the source code](https://github.com/MIna-Maher/k8s-diff-informer)

> The chart defaults to the public <code>1.1.0</code> container image, available for Linux <code>amd64</code> and <code>arm64</code>. Maintainers can follow the [release guide](https://github.com/MIna-Maher/k8s-diff-informer/blob/main/docs/RELEASING.md).

## Resource change diff example

A deployment image change produces a Slack message with the resource identity, cluster name, and image diff:

~~~text
Resource Updated — Deployment payments, namespace production
Cluster: prod-eu
- image: payments:v4
+ image: payments:v5
~~~

The minus sign marks the previous value; the plus sign marks the new value. Configured ignored field paths are removed before comparison, so they do not appear in the diff. Configured redacted paths remain visible as changed paths with values masked. Secret `data` and `stringData` values and common credential-like fields are redacted automatically. The application skips initial informer contents and sends updates only when a remaining field differs.

## Interface

The screenshots use generic demo values and show resource additions, a field-level update, and a deletion:

### Resource added

New-resource notifications identify the Kubernetes kind, name, namespace, and cluster.

![Slack resource added message]({{ '/assets/slack-resource-added.jpg' | relative_url }})

### Resource updated

Only changed fields that remain after ignored paths are removed appear in the resource change diff.

![Slack resource change diff]({{ '/assets/slack-resource-updated.jpg' | relative_url }})

### Resource deleted

Deletion notifications identify the removed resource and its namespace.

![Slack resource deleted message]({{ '/assets/slack-resource-deleted.jpg' | relative_url }})

## Requirements and limits

- Run one replica. Multiple watchers can duplicate notifications.
- The queue is in memory; pending work can be lost during a restart.
- Update diffs are sent to Slack. Secret values, common credential-like fields, and explicitly configured redacted paths are masked. Automatic checks cannot detect every secret embedded in arbitrary strings; configure sensitive paths and restrict watched resources and Slack access.

This site is built and deployed with GitHub Pages from the <code>site/</code> directory.
