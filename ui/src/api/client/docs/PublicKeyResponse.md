# PublicKeyResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | **string** | Public key\&#39;s data.  The &#x60;data&#x60; field receives the public key enconded as &#x60;base64&#x60; string.  | [optional] [default to undefined]
**fingerprint** | **string** | Public key\&#39;s fingerprint. | [optional] [default to undefined]
**created_at** | **string** | Public key\&#39;s creation date. | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**name** | **string** | Public key\&#39;s name. | [optional] [default to undefined]
**filter** | [**PublicKeyFilter**](PublicKeyFilter.md) |  | [optional] [default to undefined]
**username** | **string** | Public key\&#39;s regex username.   The &#x60;username&#x60; field define which user, in the device, may be access through this public key.  | [optional] [default to undefined]

## Example

```typescript
import { PublicKeyResponse } from './api';

const instance: PublicKeyResponse = {
    data,
    fingerprint,
    created_at,
    tenant_id,
    name,
    filter,
    username,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
