# InfoAuthentication

Authentication methods available for the ShellHub instance.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**local** | **boolean** | Indicates if local authentication using email and password is enabled. | [optional] [default to undefined]
**saml** | **boolean** | Indicates if SAML-based single sign-on (SSO) is enabled. | [optional] [default to undefined]

## Example

```typescript
import { InfoAuthentication } from './api';

const instance: InfoAuthentication = {
    local,
    saml,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
