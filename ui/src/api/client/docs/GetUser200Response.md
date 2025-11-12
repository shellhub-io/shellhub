# GetUser200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | User\&#39;s ID. | [default to undefined]
**status** | [**GetUser200ResponseStatus**](GetUser200ResponseStatus.md) |  | [default to undefined]
**max_namespaces** | **number** | Maximum number of namespaces the user can own | [default to undefined]
**created_at** | **string** | User\&#39;s creating date | [default to undefined]
**last_login** | **string** | User\&#39;s last login date | [default to undefined]
**name** | **string** | User\&#39;s name. | [default to undefined]
**username** | **string** | User\&#39;s username. | [default to undefined]
**email** | **string** | User\&#39;s E-mail. | [default to undefined]
**recovery_email** | **string** | User\&#39;s recovery email address | [optional] [default to undefined]
**mfa** | [**GetUser200ResponseMfa**](GetUser200ResponseMfa.md) |  | [default to undefined]
**namespacesOwned** | **number** | Number of namespaces owned by the user | [default to undefined]
**preferences** | [**GetUser200ResponsePreferences**](GetUser200ResponsePreferences.md) |  | [default to undefined]

## Example

```typescript
import { GetUser200Response } from './api';

const instance: GetUser200Response = {
    id,
    status,
    max_namespaces,
    created_at,
    last_login,
    name,
    username,
    email,
    recovery_email,
    mfa,
    namespacesOwned,
    preferences,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
