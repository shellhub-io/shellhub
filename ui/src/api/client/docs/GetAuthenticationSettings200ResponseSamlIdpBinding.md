# GetAuthenticationSettings200ResponseSamlIdpBinding

Configuration for SAML binding methods that define how authentication requests  and responses are transmitted between ShellHub and the IdP. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**post** | **string** | The Single Sign-On URL for HTTP-POST binding. This URL is where ShellHub  will redirect users for authentication using the HTTP-POST method.  | [optional] [default to undefined]
**redirect** | **string** | The Single Sign-On URL for HTTP-Redirect binding. This URL is where ShellHub  will redirect users for authentication using the HTTP-Redirect method.  | [optional] [default to undefined]
**preferred** | [**GetAuthenticationSettings200ResponseSamlIdpBindingPreferred**](GetAuthenticationSettings200ResponseSamlIdpBindingPreferred.md) |  | [optional] [default to undefined]

## Example

```typescript
import { GetAuthenticationSettings200ResponseSamlIdpBinding } from './api';

const instance: GetAuthenticationSettings200ResponseSamlIdpBinding = {
    post,
    redirect,
    preferred,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
