# UpdateRecoverPasswordRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**password** | **string** | User\&#39;s password. | [default to undefined]
**token** | **string** | User\&#39;s recovery token.    It is the token from the email sent to user when the user request password reset.  | [default to undefined]

## Example

```typescript
import { UpdateRecoverPasswordRequest } from './api';

const instance: UpdateRecoverPasswordRequest = {
    password,
    token,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
