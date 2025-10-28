# ConfigureSAMLAuthenticationRequestIdpMappings

Defines how SAML attributes from the IdP should be mapped to ShellHub user attributes. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**email** | **string** | The name of the SAML attribute that contains the user\&#39;s email address. This attribute will be used to identify and authenticate users in ShellHub. For example, if your IdP sends the email in a SAML attribute named \&quot;mail\&quot;, set this value to \&quot;mail\&quot;.  | [optional] [default to 'emailAddress']
**name** | **string** | The name of the SAML attribute that contains the user\&#39;s display name. This attribute will be used as the user\&#39;s name in ShellHub. For example, if your IdP sends the user\&#39;s name in a SAML attribute  named \&quot;displayName\&quot;, set this value to \&quot;displayName\&quot;.  | [optional] [default to 'displayName']

## Example

```typescript
import { ConfigureSAMLAuthenticationRequestIdpMappings } from './api';

const instance: ConfigureSAMLAuthenticationRequestIdpMappings = {
    email,
    name,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
