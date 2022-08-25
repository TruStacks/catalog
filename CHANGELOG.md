## 2.4.0 (2022-08-24)

### Feat

- add cert manager support to argo cd. add cert manager and ingress class parameters. (#16)

## 2.3.0 (2022-08-24)

### Feat

- add resource limits to concourse tasks (#15)

## 2.2.0 (2022-08-16)

### Feat

- change the concourse obfuscation format

## 2.1.0 (2022-08-11)

### Feat

- obfuscate concourse secrets (#13)

## 2.0.6 (2022-08-09)

### Fix

- remove admins role whitespace (#12)

## 2.0.5 (2022-08-08)

### Fix

- change the rbac string (#11)

## 2.0.4 (2022-08-07)

### Fix

- update argocd trustacks user rbac permissions (#10)

## 2.0.3 (2022-08-06)

### Fix

- change get application input args position (#9)

## 2.0.2 (2022-08-06)

### Fix

- change sops ci args position (#8)

## 2.0.1 (2022-08-05)

### Fix

- remove hook annotations from the concourse-web role and binding. add success hook annotation to argocd (#7)

## 2.0.0 (2022-08-05)

### Feat

- refactor concourse ci driver application creation. update catalog parameters (#5)

## 1.1.0 (2022-08-04)

### Feat

- add concourse application hooks (#4)

## 1.0.2 (2022-07-26)

### Fix

- change the catalog hook source version tag (#1)

## 1.0.1 (2022-07-26)

### Fix

- move the catalog version to the catalog hook source

## 1.0.0 (2022-07-26)

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
