# ConnectorInfo200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ID** | **string** |  | [optional] [default to undefined]
**Containers** | **number** |  | [optional] [default to undefined]
**ContainersRunning** | **number** |  | [optional] [default to undefined]
**ContainersPaused** | **number** |  | [optional] [default to undefined]
**ContainersStopped** | **number** |  | [optional] [default to undefined]
**Images** | **number** |  | [optional] [default to undefined]
**Driver** | **string** |  | [optional] [default to undefined]
**DriverStatus** | **Array&lt;Array&lt;string&gt;&gt;** |  | [optional] [default to undefined]
**Plugins** | [**ConnectorInfo200ResponsePlugins**](ConnectorInfo200ResponsePlugins.md) |  | [optional] [default to undefined]
**MemoryLimit** | **boolean** |  | [optional] [default to undefined]
**SwapLimit** | **boolean** |  | [optional] [default to undefined]
**CpuCfsPeriod** | **boolean** |  | [optional] [default to undefined]
**CpuCfsQuota** | **boolean** |  | [optional] [default to undefined]
**CPUShares** | **boolean** |  | [optional] [default to undefined]
**CPUSet** | **boolean** |  | [optional] [default to undefined]
**PidsLimit** | **boolean** |  | [optional] [default to undefined]
**IPv4Forwarding** | **boolean** |  | [optional] [default to undefined]
**BridgeNfIptables** | **boolean** |  | [optional] [default to undefined]
**BridgeNfIp6tables** | **boolean** |  | [optional] [default to undefined]
**Debug** | **boolean** |  | [optional] [default to undefined]
**NFd** | **number** |  | [optional] [default to undefined]
**OomKillDisable** | **boolean** |  | [optional] [default to undefined]
**NGoroutines** | **number** |  | [optional] [default to undefined]
**SystemTime** | **string** |  | [optional] [default to undefined]
**LoggingDriver** | **string** |  | [optional] [default to undefined]
**CgroupDriver** | **string** |  | [optional] [default to undefined]
**CgroupVersion** | **string** |  | [optional] [default to undefined]
**NEventsListener** | **number** |  | [optional] [default to undefined]
**KernelVersion** | **string** |  | [optional] [default to undefined]
**OperatingSystem** | **string** |  | [optional] [default to undefined]
**OSVersion** | **string** |  | [optional] [default to undefined]
**OSType** | **string** |  | [optional] [default to undefined]
**Architecture** | **string** |  | [optional] [default to undefined]
**IndexServerAddress** | **string** |  | [optional] [default to undefined]
**RegistryConfig** | [**ConnectorInfo200ResponseRegistryConfig**](ConnectorInfo200ResponseRegistryConfig.md) |  | [optional] [default to undefined]
**NCPU** | **number** |  | [optional] [default to undefined]
**MemTotal** | **number** |  | [optional] [default to undefined]
**GenericResources** | **Array&lt;string | null&gt;** |  | [optional] [default to undefined]
**DockerRootDir** | **string** |  | [optional] [default to undefined]
**HttpProxy** | **string** |  | [optional] [default to undefined]
**HttpsProxy** | **string** |  | [optional] [default to undefined]
**NoProxy** | **string** |  | [optional] [default to undefined]
**Name** | **string** |  | [optional] [default to undefined]
**Labels** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**ExperimentalBuild** | **boolean** |  | [optional] [default to undefined]
**ServerVersion** | **string** |  | [optional] [default to undefined]
**Runtimes** | [**ConnectorInfo200ResponseRuntimes**](ConnectorInfo200ResponseRuntimes.md) |  | [optional] [default to undefined]
**DefaultRuntime** | **string** |  | [optional] [default to undefined]
**Swarm** | [**ConnectorInfo200ResponseSwarm**](ConnectorInfo200ResponseSwarm.md) |  | [optional] [default to undefined]
**LiveRestoreEnabled** | **boolean** |  | [optional] [default to undefined]
**Isolation** | **string** |  | [optional] [default to undefined]
**InitBinary** | **string** |  | [optional] [default to undefined]
**ContainerdCommit** | [**ConnectorInfo200ResponseContainerdCommit**](ConnectorInfo200ResponseContainerdCommit.md) |  | [optional] [default to undefined]
**RuncCommit** | [**ConnectorInfo200ResponseContainerdCommit**](ConnectorInfo200ResponseContainerdCommit.md) |  | [optional] [default to undefined]
**InitCommit** | [**ConnectorInfo200ResponseContainerdCommit**](ConnectorInfo200ResponseContainerdCommit.md) |  | [optional] [default to undefined]
**SecurityOptions** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**CDISpecDirs** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**Warnings** | **Array&lt;string&gt;** |  | [optional] [default to undefined]

## Example

```typescript
import { ConnectorInfo200Response } from './api';

const instance: ConnectorInfo200Response = {
    ID,
    Containers,
    ContainersRunning,
    ContainersPaused,
    ContainersStopped,
    Images,
    Driver,
    DriverStatus,
    Plugins,
    MemoryLimit,
    SwapLimit,
    CpuCfsPeriod,
    CpuCfsQuota,
    CPUShares,
    CPUSet,
    PidsLimit,
    IPv4Forwarding,
    BridgeNfIptables,
    BridgeNfIp6tables,
    Debug,
    NFd,
    OomKillDisable,
    NGoroutines,
    SystemTime,
    LoggingDriver,
    CgroupDriver,
    CgroupVersion,
    NEventsListener,
    KernelVersion,
    OperatingSystem,
    OSVersion,
    OSType,
    Architecture,
    IndexServerAddress,
    RegistryConfig,
    NCPU,
    MemTotal,
    GenericResources,
    DockerRootDir,
    HttpProxy,
    HttpsProxy,
    NoProxy,
    Name,
    Labels,
    ExperimentalBuild,
    ServerVersion,
    Runtimes,
    DefaultRuntime,
    Swarm,
    LiveRestoreEnabled,
    Isolation,
    InitBinary,
    ContainerdCommit,
    RuncCommit,
    InitCommit,
    SecurityOptions,
    CDISpecDirs,
    Warnings,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
