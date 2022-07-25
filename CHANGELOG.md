## 1.0.0 (2022-07-25)

### Feat

- add version to catalog. add commitizen to actions
- add trustack system user to concourse
- add cleanup hook annotations
- add argo cd service account creation
- add argo-cd hooks
- refactor functions. add concourse pre install hook
- add env var service url. update README
- add health check timeout
- add manifests. remove hooks
- remove unused components. merge authentik hooks sources. convert config imports to embeds
- add authentik component. add server functions
- add argocd component
- add fullname override
- remnove buildkit sidecar
- add concourse secret variables
- change the service port
- refactor modules. add catalog configuration
- add unit tests to components
- update component values
- update values datatype. change pkg import
- convert values to string
- change default ingress port to string
- add sealedsecrets
- add sealed secrets component
- refactor components
- add mask to password parameter
- change the components path
- add component parameters
- add hook source to catalog
- add version to baseComponent interface methods
- add README.md. add Dockerfile. update components.

### Refactor

- update concourse docs and variables

### Fix

- update concourse username claim. add signing key to oidc client
- add missing sources
- update hook namespaces. add service health check to post install hook
- add the hook resources and weights. add postgresql fullname overrides
- add authentik hooks
- change authentik hook names. add hook name constants
- resolve authentik indentation issues
- indent authentik bootstrap token env var
- add equal comparison to concourse template tls
- update the authentik postgresql password template
- update value variables
- use current working directory path to assets
- add idempotent /data creation
- lint errors. add asset linking to action workflow
- change the sealed secrets helm chart source
- update Dockerfile
- update text fixture
- change the externalURL template to multiline
- change the concourse values template to valid yaml
- update concourse values
- remove pre-install hook
- convert env var to string
- add components to includes
- resolve lint errors
- update helm repo roots
- update helm chart names
- resolve lint errors
