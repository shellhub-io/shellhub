# Tunnel


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**address** | **string** | Tunnel\&#39;s unique address | [optional] [default to undefined]
**full_address** | **string** | Full tunnel address including domain | [optional] [default to undefined]
**namespace** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**device** | **string** | Device\&#39;s UID | [optional] [default to undefined]
**host** | **string** | Tunnel\&#39;s agent host address | [optional] [default to undefined]
**port** | **number** | Tunnel\&#39;s agent port number | [optional] [default to undefined]
**ttl** | **number** | Tunnel\&#39;s time to live in seconds | [optional] [default to undefined]
**expires_in** | **string** | Tunnel\&#39;s expiration date | [optional] [default to undefined]
**created_at** | **string** | Tunnel\&#39;s creation date | [optional] [default to undefined]

## Example

```typescript
import { Tunnel } from './api';

const instance: Tunnel = {
    address,
    full_address,
    namespace,
    device,
    host,
    port,
    ttl,
    expires_in,
    created_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
