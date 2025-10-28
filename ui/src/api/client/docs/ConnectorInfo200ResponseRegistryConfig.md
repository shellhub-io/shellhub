# ConnectorInfo200ResponseRegistryConfig


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowNondistributableArtifactsCIDRs** | **Array&lt;string | null&gt;** |  | [optional] [default to undefined]
**AllowNondistributableArtifactsHostnames** | **Array&lt;string | null&gt;** |  | [optional] [default to undefined]
**InsecureRegistryCIDRs** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**IndexConfigs** | [**ConnectorInfo200ResponseRegistryConfigIndexConfigs**](ConnectorInfo200ResponseRegistryConfigIndexConfigs.md) |  | [optional] [default to undefined]
**Mirrors** | **Array&lt;string&gt;** |  | [optional] [default to undefined]

## Example

```typescript
import { ConnectorInfo200ResponseRegistryConfig } from './api';

const instance: ConnectorInfo200ResponseRegistryConfig = {
    AllowNondistributableArtifactsCIDRs,
    AllowNondistributableArtifactsHostnames,
    InsecureRegistryCIDRs,
    IndexConfigs,
    Mirrors,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
