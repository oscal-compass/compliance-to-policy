# Canaries

One object per probed check, each written to **fail that check and pass the others**. Before the
probe trusts a pass, it submits the canary for that check and requires the check to report it as
failing. If the canary comes back clean, the probe cannot see that check -- a selector that
matches nothing, a policy that did not compile, a `matchConstraints` that does not cover the
resource -- and every verdict for it is `error` rather than `pass`.

Without this, silence is ambiguous: a probe that is wired up wrongly reports a perfect posture.

These objects are only ever submitted as a server-side **dry run**, into a scratch namespace the
probe creates and deletes. They are never created, which is why `replicas: 0` and an unresolvable
image digest are harmless here. Nothing pulls the image and no pod is scheduled.

A check named in a component definition but missing a canary here is reported as `error`, with
the path it looked for. Add the file rather than reaching for a flag to turn the check off.
