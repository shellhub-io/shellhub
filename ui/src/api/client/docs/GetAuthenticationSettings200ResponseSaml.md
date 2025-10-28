# GetAuthenticationSettings200ResponseSaml


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enabled** | **boolean** | Indicates whether SAML authentication is active. | [optional] [default to undefined]
**auth_url** | **string** | The URL used to authenticate the SAML user. | [optional] [default to undefined]
**assertion_url** | **string** | The URL where the IdP must permit redirects. | [optional] [default to undefined]
**idp** | [**GetAuthenticationSettings200ResponseSamlIdp**](GetAuthenticationSettings200ResponseSamlIdp.md) |  | [optional] [default to undefined]
**sp** | [**GetAuthenticationSettings200ResponseSamlSp**](GetAuthenticationSettings200ResponseSamlSp.md) |  | [optional] [default to undefined]

## Example

```typescript
import { GetAuthenticationSettings200ResponseSaml } from './api';

const instance: GetAuthenticationSettings200ResponseSaml = {
    enabled,
    auth_url,
    assertion_url,
    idp,
    sp,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
