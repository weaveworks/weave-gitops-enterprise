## Custom GUI 
We support the following custom GUI elements:
- Logo (Logo Dimension should not be bigger than ( width: '150px', height: '32px')
- Footer (We do support mark-down style in the footer content but we don't encourage it as it will break the footer style.)

## Enabling

We make some configuration changes to the `values` in the Weave GitOps Enterprise `HelmRelease`.

```yaml
    logoURL: http://iqt.dev/iqt.svg
    footer:
      backgroundColor: red
      color: blue
      content: My footer header This is the footer [link](example.com)
      # Show or hide version of Weave GitOps
      hideVersion: false
```

## Rollout

After updating the `HelmRelease`, commit, push, reconcile and restart:

- `git commit -am "Enable cost estimation"`
- `git push`
- `flux reconcile kustomization --with-source flux-system`
- `kubectl -n flux-system rollout restart deploy/weave-gitops-enterprise-mccp-cluster-service`
