# PublicKeyRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**data** | **string** | Public key\&#39;s data.  The &#x60;data&#x60; field receives the public key enconded as &#x60;base64&#x60; string.  | [default to undefined]
**filter** | [**PublicKeyFilter**](PublicKeyFilter.md) |  | [default to undefined]
**name** | **string** | Public key\&#39;s name. | [default to undefined]
**username** | **string** | Public key\&#39;s regex username.   The &#x60;username&#x60; field define which user, in the device, may be access through this public key.  | [default to undefined]

## Example

```typescript
import { PublicKeyRequest } from './api';

const instance: PublicKeyRequest = {
    data,
    filter,
    name,
    username,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
