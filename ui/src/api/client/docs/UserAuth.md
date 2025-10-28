# UserAuth


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **string** | JWT Token | [optional] [default to undefined]
**id** | **string** | User\&#39;s ID. | [optional] [default to undefined]
**origin** | [**UserOrigin**](UserOrigin.md) |  | [optional] [default to undefined]
**user** | **string** | User\&#39;s username. | [optional] [default to undefined]
**name** | **string** | User\&#39;s name. | [optional] [default to undefined]
**email** | **string** | User\&#39;s E-mail. | [optional] [default to undefined]
**recovery_email** | **string** | The recovery email serves as the user\&#39;s final recourse to regain access to their account.  | [optional] [default to undefined]
**tenant** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**role** | [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] [default to undefined]
**mfa** | **boolean** | Indicates whether the user has MFA enabled. | [optional] [default to undefined]
**max_namespaces** | **number** | Indicates the maximum number of namespaces a user is allowed to create. If set to 0, the user is not permitted to create any namespaces. If set to -1, the user has no limit on the number of namespaces they can create. | [optional] [default to undefined]

## Example

```typescript
import { UserAuth } from './api';

const instance: UserAuth = {
    token,
    id,
    origin,
    user,
    name,
    email,
    recovery_email,
    tenant,
    role,
    mfa,
    max_namespaces,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
