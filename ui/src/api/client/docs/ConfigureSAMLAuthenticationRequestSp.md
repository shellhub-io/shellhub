# ConfigureSAMLAuthenticationRequestSp

Configuration settings for how ShellHub will function as a Service Provider (SP).

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**sign_requests** | **boolean** | Indicates whether ShellHub should sign authentication requests.  If &#x60;true&#x60;, ShellHub will generate a public X509 certificate that must be deployed on the IdP for validation.  The IdP-initiated workflow is not supported when this option is enabled.  | [optional] [default to undefined]

## Example

```typescript
import { ConfigureSAMLAuthenticationRequestSp } from './api';

const instance: ConfigureSAMLAuthenticationRequestSp = {
    sign_requests,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
