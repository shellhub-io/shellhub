# UpdateUserDataRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | User\&#39;s name. | [optional] [default to undefined]
**username** | **string** | User\&#39;s username. | [optional] [default to undefined]
**email** | **string** | User\&#39;s e-mail. | [optional] [default to undefined]
**recovery_email** | **string** | User\&#39;s recovery e-mail. A recovery email serves as the user\&#39;s final recourse to regain access to their account.  | [optional] [default to undefined]

## Example

```typescript
import { UpdateUserDataRequest } from './api';

const instance: UpdateUserDataRequest = {
    name,
    username,
    email,
    recovery_email,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
