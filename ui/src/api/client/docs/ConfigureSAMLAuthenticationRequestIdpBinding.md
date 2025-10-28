# ConfigureSAMLAuthenticationRequestIdpBinding

Configuration for SAML binding methods that define how authentication requests  and responses are transmitted between ShellHub and the IdP. SAML supports  different binding protocols for flexibility in deployment scenarios. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**post** | **string** | The Single Sign-On URL for HTTP-POST binding. This URL is where ShellHub  will redirect users for authentication using the HTTP-POST method, which  sends SAML data in the body of an HTTP POST request. This binding is more  secure as it doesn\&#39;t expose SAML data in URL parameters.  | [optional] [default to undefined]
**redirect** | **string** | The Single Sign-On URL for HTTP-Redirect binding. This URL is where ShellHub  will redirect users for authentication using the HTTP-Redirect method, which  sends SAML data as URL parameters. This binding is simpler but has URL length  limitations and exposes SAML data in browser history and server logs.  | [optional] [default to undefined]
**preferred** | [**ConfigureSAMLAuthenticationRequestIdpBindingPreferred**](ConfigureSAMLAuthenticationRequestIdpBindingPreferred.md) |  | [optional] [default to undefined]

## Example

```typescript
import { ConfigureSAMLAuthenticationRequestIdpBinding } from './api';

const instance: ConfigureSAMLAuthenticationRequestIdpBinding = {
    post,
    redirect,
    preferred,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
