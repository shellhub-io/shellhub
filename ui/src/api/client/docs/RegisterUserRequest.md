# RegisterUserRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | The full name of the user. | [default to undefined]
**email** | **string** | The user\&#39;s email address, which must be unique. This email will be used for login and for receiving important notifications, such as password reset emails. If &#x60;email_marketing&#x60; is set to &#x60;true&#x60;, promotional emails will also be sent to this address.  | [default to undefined]
**username** | **string** | The username, which must be unique across the system. Users can log in using either their username or email.  | [default to undefined]
**password** | **string** | The password for the user account. Must follow the regex.  | [default to undefined]
**email_marketing** | **boolean** | Indicates whether the user opts to receive marketing and promotional emails.  | [default to undefined]
**sig** | **string** | **For standard registration processes, this field should be ignored.**   A unique signature included in an invitation email. This is used to automatically confirm the user\&#39;s registration without requiring an additional confirmation email.  | [optional] [default to undefined]

## Example

```typescript
import { RegisterUserRequest } from './api';

const instance: RegisterUserRequest = {
    name,
    email,
    username,
    password,
    email_marketing,
    sig,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
