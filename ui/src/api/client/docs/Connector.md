# Connector


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**uid** | **string** | Connector\&#39;s UID | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**enable** | **boolean** | Connector\&#39;s connection is enabled. | [optional] [default to undefined]
**address** | **string** | Address to the Container Engine. | [optional] [default to undefined]
**port** | **number** | Port to the Container Engine. | [optional] [default to undefined]
**secure** | **boolean** | onnector\&#39;s connection is using HTTPS for authentication. | [optional] [default to undefined]
**status** | [**ConnectorStatus**](ConnectorStatus.md) |  | [optional] [default to undefined]
**tls** | [**ConnectorTLS**](ConnectorTLS.md) |  | [optional] [default to undefined]

## Example

```typescript
import { Connector } from './api';

const instance: Connector = {
    uid,
    tenant_id,
    enable,
    address,
    port,
    secure,
    status,
    tls,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
