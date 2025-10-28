# UserAdminRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | User\&#39;s name. | [default to undefined]
**email** | **string** | User\&#39;s E-mail. | [default to undefined]
**username** | **string** | User\&#39;s username. | [default to undefined]
**password** | **string** | User\&#39;s password. | [default to undefined]
**confirmed** | **boolean** | User\&#39;s email confirmed. | [optional] [default to false]
**max_namespaces** | **number** | Indicates the maximum number of namespaces a user is allowed to create. If set to 0, the user is not permitted to create any namespaces. If set to -1, the user has no limit on the number of namespaces they can create. | [optional] [default to undefined]

## Example

```typescript
import { UserAdminRequest } from './api';

const instance: UserAdminRequest = {
    name,
    email,
    username,
    password,
    confirmed,
    max_namespaces,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
