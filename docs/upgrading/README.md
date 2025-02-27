<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Upgrading Instruction](#upgrading-instruction)
  - [Overview](#overview)
  - [Regular Upgrading Process](#regular-upgrading-process)
    - [Upgrading APIs](#upgrading-apis)
      - [Manual Upgrade API](#manual-upgrade-api)
    - [Upgrading Components](#upgrading-components)
  - [Details Upgrading Instruction](#details-upgrading-instruction)
    - [v0.8 to v0.9](#v08-to-v09)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# Upgrading Instruction

## Overview
Karmada uses the [semver versioning](https://semver.org/) and each version in the format of v`MAJOR`.`MINOR`.`PATCH`:
- The `PATCH` release does not introduce breaking changes.
- The `MINOR` release might introduce minor breaking changes with a workaround.
- The `Major` release might introduce backward-incompatible behavior changes.

## Regular Upgrading Process
### Upgrading APIs
For releases that introduce API changes, the Karmada API(CRD) that Karmada components rely on must upgrade to keep consistent.

Karmada CRD is composed of two parts:
- bases: The CRD definition generated via API structs.
- patches: conversion settings for the CRD.

In order to support multiple versions of custom resources, the `patches` should be injected into `bases`.
To achieve this we introduced a `kustomization.yaml` configuration then use `kubectl kustomize` to build the final CRD.

The `bases`,`patches`and `kustomization.yaml` now located at `charts/_crds` directory of the repo.

#### Manual Upgrade API

**Step 1: Get the Webhook CA certificate**
The CA certificate will be injected into `patches` before building the final CRD.
We can retrieve it from the `MutatingWebhookConfiguration` or `ValidatingWebhookConfiguration` configurations, e.g:
```bash
kubectl get mutatingwebhookconfigurations.admissionregistration.k8s.io mutating-config
```
Copy the `ca_string` from the yaml path `webhooks.name[x].clientConfig.caBundle`, then replace the `{{caBundle}}` from
the yaml files in `patches`. e.g:
```bash
sed -i'' -e "s/{{caBundle}}/${ca_string}/g" ./"charts/_crds/patches/webhook_in_resourcebindings.yaml"
sed -i'' -e "s/{{caBundle}}/${ca_string}/g" ./"charts/_crds/patches/webhook_in_clusterresourcebindings.yaml"
```

**Step2: Build final CRD**
Generate the final CRD by `kubectl kustomize` command, e.g:
```bash
kubectl kustomize ./charts/_crds 
```
Or, you can apply to `karmada-apiserver` by:
```bash
kubectl kustomize ./charts/_crds | kubectl apply -f -
```

### Upgrading Components
Components upgrading is composed of image version update and possible command args changes.

> For the argument changes please refer to `Details Upgrading Instruction` below.

## Details Upgrading Instruction
### [v0.8 to v0.9](./v0.8-v0.9.md)
