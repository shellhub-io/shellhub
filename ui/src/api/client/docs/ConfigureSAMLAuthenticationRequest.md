# ConfigureSAMLAuthenticationRequest


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enable** | **boolean** | Specifies whether to activate (&#x60;true&#x60;) or deactivate (&#x60;false&#x60;) SAML authentication. If set to &#x60;false&#x60;, all other attributes will be ignored.  | [default to undefined]
**idp** | [**ConfigureSAMLAuthenticationRequestIdp**](ConfigureSAMLAuthenticationRequestIdp.md) |  | [default to undefined]
**sp** | [**ConfigureSAMLAuthenticationRequestSp**](ConfigureSAMLAuthenticationRequestSp.md) |  | [default to undefined]

## Example

```typescript
import { ConfigureSAMLAuthenticationRequest } from './api';

const instance: ConfigureSAMLAuthenticationRequest = {
    enable,
    idp,
    sp,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
