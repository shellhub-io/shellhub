# UpdateUserRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** |  | [optional] [default to undefined]
**username** | **string** |  | [optional] [default to undefined]
**email** | **string** |  | [optional] [default to undefined]
**recovery_email** | **string** | A recovery email serves as the user\&#39;s final recourse to regain access to their account. It cannot be the same as the user\&#39;s primary email. Once defined, it cannot be updated to an empty value.  | [optional] [default to undefined]
**password** | **string** |  | [optional] [default to undefined]
**current_password** | **string** | It\&#39;s required when updating the user\&#39;s password.  | [optional] [default to undefined]

## Example

```typescript
import { UpdateUserRequest } from './api';

const instance: UpdateUserRequest = {
    name,
    username,
    email,
    recovery_email,
    password,
    current_password,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
