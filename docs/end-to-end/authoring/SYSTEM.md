# The system being documented

A hand-rolled three-tier storefront on a vanilla Kubernetes cluster: a `web` tier, an `api`
tier and a `db` tier, all in namespace `storefront`. No service mesh, no operators, no custom
resources of our own. Kubernetes 1.34; Kyverno 1.19.2 is installed.

Container images are built in CI, signed with Sigstore cosign and pushed to a registry. The set
of registries the cluster accepts is held in the cluster itself: one ConfigMap per registry in
namespace `supply-chain`, each labelled `supply-chain.example.com/ledger=approved-registry`,
with the registry host in its `data.registry`. The organisation's approved public key is
published as the `cosign.pub` entry of ConfigMap `storefront-cosign-approved` in the same
namespace.

Enforcement is at admission time, by policy engines. There is no node-level agent, no firmware
management and no hardware inventory in scope: an admission policy sees pod specs, not nodes.

What the organisation requires of this system:
- every image carries a valid cosign signature, from the release workflow identity or from the
  organisation's own key
- every image comes from a registry listed in the approved-registry ledger
- every image is referenced by immutable digest, never by a mutable tag
- no container runs privileged or may escalate privileges; capabilities are dropped; it runs as
  non-root; no host namespaces and no hostPath volumes
- every workload carries inventory labels: application name, part-of, data classification,
  owning team
- database credentials reach the db and api tiers only through secretKeyRef, never as a literal
  value in the manifest
