# UserAdminResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | User\&#39;s ID. | [optional] [default to undefined]
**namespaces** | **number** | User\&#39;s integer of owned namespaces. | [optional] [default to undefined]
**confirmed** | **boolean** | User\&#39;s confirmation. | [optional] [default to undefined]
**created_at** | **string** | User\&#39;s creating date. | [optional] [default to undefined]
**last_login** | **string** | User\&#39;s last login date. | [optional] [default to undefined]
**name** | **string** | User\&#39;s name. | [optional] [default to undefined]
**email** | **string** | User\&#39;s E-mail. | [optional] [default to undefined]
**username** | **string** | User\&#39;s username. | [optional] [default to undefined]
**password** | **string** | User\&#39;s hashed password. | [optional] [default to undefined]

## Example

```typescript
import { UserAdminResponse } from './api';

const instance: UserAdminResponse = {
    id,
    namespaces,
    confirmed,
    created_at,
    last_login,
    name,
    email,
    username,
    password,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
