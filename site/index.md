---
layout: default
title: k8s-diff-informer
description: Watch Kubernetes resources and send meaningful changes to Slack.
---

# k8s-diff-informer

Watch selected Kubernetes resources, compare changes, and send Slack notifications.

This project provides a dynamic Kubernetes informer, configurable field filtering, an in-memory notification queue, and Prometheus metrics.

![Architecture and event flow]({{ '/assets/architecture.svg' | relative_url }})

## Get started

- [Install with Helm]({{ '/install.html' | relative_url }})
- [Configure monitoring]({{ '/monitoring.html' | relative_url }})
- [Read the repository README](https://github.com/MIna-Maher/k8s-diff-informer#readme)
- [View the source code](https://github.com/MIna-Maher/k8s-diff-informer)

> The chart defaults to the <code>1.0.0</code> container image. The stable image has not been published yet. Build and push an image from the repository, then override <code>image.repository</code> and <code>image.tag</code> in the Helm install command.

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
