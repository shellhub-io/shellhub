# ApiKeyUpdate


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | The name of the API key. This serves as an \&quot;external ID\&quot; since the UUID will never be returned. It is unique per namespace.  | [optional] [default to undefined]
**role** | **string** | The role of the key. It serves as a \&quot;level\&quot; indicating which endpoints the key can access. It must be less or equal than the user\&#39;s role.  | [optional] [default to undefined]

## Example

```typescript
import { ApiKeyUpdate } from './api';

const instance: ApiKeyUpdate = {
    name,
    role,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
