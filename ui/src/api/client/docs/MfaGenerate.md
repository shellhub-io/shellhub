# MfaGenerate


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**link** | **string** | The link to establish a connection with the OTP server. | [optional] [default to undefined]
**secret** | **string** | A secret key to authenticate with the OTP server. | [optional] [default to undefined]
**recovery_codes** | **Array&lt;string&gt;** | A list of recovery codes to use when the user loses access to their MFA app. | [optional] [default to undefined]

## Example

```typescript
import { MfaGenerate } from './api';

const instance: MfaGenerate = {
    link,
    secret,
    recovery_codes,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
