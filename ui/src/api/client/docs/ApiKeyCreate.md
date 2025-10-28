# ApiKeyCreate


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | The name of the API key. This serves as an \&quot;external ID\&quot; since the UUID will never be returned. It is unique per namespace.  | [default to undefined]
**expires_at** | [**ApiKeyCreateExpiresAt**](ApiKeyCreateExpiresAt.md) |  | [default to undefined]
**role** | **string** | The role of the key. It serves as a \&quot;level\&quot; indicating which endpoints the key can access. It must be less or equal than the user\&#39;s role. Leave it blank to use the user\&#39;s role.  | [optional] [default to undefined]
**key** | **string** | An optional and unique value to be used as the API key\&#39;s internal identifier. This value is the \&quot;internal ID\&quot; and will NEVER be returned to the client. Leave it blank for a random one to be generated.  | [optional] [default to undefined]

## Example

```typescript
import { ApiKeyCreate } from './api';

const instance: ApiKeyCreate = {
    name,
    expires_at,
    role,
    key,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
