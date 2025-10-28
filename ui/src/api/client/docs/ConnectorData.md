# ConnectorData


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enable** | **boolean** | Connector\&#39;s connection is enabled. | [optional] [default to undefined]
**address** | **string** | Address to the Container Engine. | [optional] [default to undefined]
**port** | **number** | Port to the Container Engine. | [optional] [default to undefined]
**secure** | **boolean** | onnector\&#39;s connection is using HTTPS for authentication. | [optional] [default to undefined]
**tls** | [**ConnectorTLS**](ConnectorTLS.md) |  | [optional] [default to undefined]

## Example

```typescript
import { ConnectorData } from './api';

const instance: ConnectorData = {
    enable,
    address,
    port,
    secure,
    tls,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
