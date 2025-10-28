# ConfigureSAMLAuthenticationRequestIdp

Configuration for the Identity Provider (IdP) that ShellHub will use for authentication.  You can provide either a metadata URL to fetch all necessary information automatically  or specify the `SignOn URL`, `Entity ID`, and `X509 certificate` explicitly. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**metadata_url** | **string** | The URL to fetch IdP metadata. If provided, this will automatically retrieve all necessary configuration details and take precedence over manually specified values.  | [optional] [default to undefined]
**entity_id** | **string** | The Entity ID of the IdP. | [optional] [default to undefined]
**certificate** | **string** | The public X509 certificate of the IdP. It can be provided with or without  the PEM delimiters (&#x60;-----BEGIN CERTIFICATE-----&#x60; and &#x60;-----END CERTIFICATE-----&#x60;).  | [optional] [default to undefined]
**binding** | [**ConfigureSAMLAuthenticationRequestIdpBinding**](ConfigureSAMLAuthenticationRequestIdpBinding.md) |  | [optional] [default to undefined]
**mappings** | [**ConfigureSAMLAuthenticationRequestIdpMappings**](ConfigureSAMLAuthenticationRequestIdpMappings.md) |  | [optional] [default to undefined]

## Example

```typescript
import { ConfigureSAMLAuthenticationRequestIdp } from './api';

const instance: ConfigureSAMLAuthenticationRequestIdp = {
    metadata_url,
    entity_id,
    certificate,
    binding,
    mappings,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
