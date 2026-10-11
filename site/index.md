---
layout: default
title: k8s-diff-informer
description: A focused Kubernetes change feed: watch selected resources, filter noisy fields, and get readable Slack diffs for meaningful changes.
---

# k8s-diff-informer

<p><img src="{{ '/assets/logo.svg' | relative_url }}" alt="k8s-diff-informer logo" width="96"></p>

**A focused change feed for Kubernetes.** See what changed without comparing YAML by hand: choose the resources to watch, ignore noisy fields, and get readable Slack diffs for additions, meaningful updates, and deletions. Built for platform and Kubernetes operators who want a lightweight view of object-level changes.

The project answers “what changed in this resource?” Use your monitoring stack for cluster health and condition-based alerts. The Go service includes retrying Slack delivery, Prometheus metrics, and health endpoints.

![Architecture and event flow]({{ '/assets/architecture.svg' | relative_url }})

## Get started

- [Install with Helm]({{ '/install.html' | relative_url }})
- [Configure monitoring]({{ '/monitoring.html' | relative_url }})
- [Read the repository README](https://github.com/MIna-Maher/k8s-diff-informer#readme)
- [View the source code](https://github.com/MIna-Maher/k8s-diff-informer)

> The chart defaults to the public <code>1.0.0</code> container image, available for Linux <code>amd64</code> and <code>arm64</code>. Maintainers can follow the [release guide](https://github.com/MIna-Maher/k8s-diff-informer/blob/main/docs/RELEASING.md).

## Notification example

A deployment image change produces a Slack message with the resource identity, cluster name, and image diff:

~~~text
Resource Updated — Deployment payments, namespace production
Cluster: prod-eu
- image: payments:v4
+ image: payments:v5
~~~

The application skips initial informer contents, sends later add and delete notifications, and sends update notifications only when selected fields differ.

## Requirements and limits

- Run one replica. Multiple watchers can duplicate notifications.
- The queue is in memory; pending work can be lost during a restart.
- Resource diffs can contain sensitive values. Choose watched resources and ignored fields carefully.

This site is built and deployed with GitHub Pages from the <code>site/</code> directory.
