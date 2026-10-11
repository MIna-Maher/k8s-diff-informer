---
layout: default
title: k8s-diff-informer
description: Track selected Kubernetes resource changes, filter noisy fields, and send readable diffs to Slack with Prometheus monitoring.
---

# k8s-diff-informer

<p><img src="{{ '/assets/logo.svg' | relative_url }}" alt="k8s-diff-informer logo" width="96"></p>

Track meaningful changes across your Kubernetes resources. Choose what to watch, ignore fields that create noise, and receive clear Slack notifications with readable diffs when resources are added, changed, or deleted. Prometheus metrics and health endpoints are included.

A focused Go service for live Kubernetes change visibility, with configurable watches, noise filtering, retrying Slack delivery, and Prometheus monitoring.

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
