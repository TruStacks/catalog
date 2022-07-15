# Catalog

The catalog is used for the discovery of [Softwre Delivery Toolchain](https://github.com/TruStacks/factories) components.

## Components

Components are software services that are used in the software delivery toolchain. Each component must have a *helm repository*, a *helm chart*, *helm version*, and *hooks*.

The component must be tested to verify the integrity of any implemented hook functions with the specified *helm version*.

## Runtime Modes

The catalog can be run in two modes. This architecture allows the catalog to be used for both component discovery and hook execution without the need to manage additional repositories and containers. The mode is specified using the **CATALOG_MODE** environment variable. 

### server

`server` mode starts the catalog as a webserver with a route to the component manifest.

The manifest can be accessed at the path `/.well-known/catalog-manifest`.

### hook

`hook` mode starts the catalog in [helm hook](https://helm.sh/docs/topics/charts_hooks/) execution mode. The hook that will be executed is defined using two environment variables. 

**HOOK_COMPONENT** is the name of the component to run the hook against (ie. sonarqube). **HOOK_KIND** is the [type of hook](https://helm.sh/docs/topics/charts_hooks/#the-available-hooks)  to execute. 

The desired hook must implemented for the provided component. The [Base Component](https://github.com/TruStacks/catalog/blob/main/component.go) provides and 

*Server mode is the default mode if the **CATALOG_MODE** environment variable is not set.*

## Parameters

Configuration parameters are defined in the toolchain install config. The provided parameters are used in the component values. The available parameters are defined in [catalog.yaml](https://github.com/TruStacks/catalog/blob/main/pkg/catalog/catalog.yaml).

## Chart dependencies

Toolchains that use this catalog must provide the following helm chart dependencies:

| name | source | version |
| - | - | - |
| common | https://library-charts.k8s-at-home.com | >= 4.2.0