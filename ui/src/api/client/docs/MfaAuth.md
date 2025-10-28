# MfaAuth


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**token** | **string** | The &#x60;X-MFA-Token&#x60; header returned by the authUser endpoint. | [default to undefined]
**code** | **string** | The current code from the MFA authenticator. | [default to undefined]

## Example

```typescript
import { MfaAuth } from './api';

const instance: MfaAuth = {
    token,
    code,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
