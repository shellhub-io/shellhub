# ApiKey


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**tenant_id** | **string** | The tenant ID of the namespace with which the key is associated. | [optional] [default to undefined]
**created_by** | **string** | The ID of the user who created the API key. | [optional] [default to undefined]
**role** | **string** | The role of the key. It serves as a \&quot;level\&quot; indicating which endpoints the key can access.  | [optional] [default to undefined]
**name** | **string** | The name of the API key. This serves as an \&quot;external ID\&quot; since the UUID will never be returned. It is unique per namespace.  | [optional] [default to undefined]
**expires_in** | **number** | Epoch time until expiration. It for unlimited keys.  | [optional] [default to undefined]
**created_at** | **string** | The UTC date when the key was created. | [optional] [default to undefined]
**updated_at** | **string** | The UTC date when the key was last updated. It is updated whenever the key is modified.  | [optional] [default to undefined]

## Example

```typescript
import { ApiKey } from './api';

const instance: ApiKey = {
    tenant_id,
    created_by,
    role,
    name,
    expires_in,
    created_at,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
