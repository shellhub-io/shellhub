# GetAuthenticationSettings200ResponseSamlIdp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**entity_id** | **string** | The Entity ID of the IdP. | [optional] [default to undefined]
**certificates** | **Array&lt;string&gt;** | The list of public X509 certificates of the IdP. | [optional] [default to undefined]
**binding** | [**GetAuthenticationSettings200ResponseSamlIdpBinding**](GetAuthenticationSettings200ResponseSamlIdpBinding.md) |  | [optional] [default to undefined]
**mappings** | [**GetAuthenticationSettings200ResponseSamlIdpMappings**](GetAuthenticationSettings200ResponseSamlIdpMappings.md) |  | [optional] [default to undefined]

## Example

```typescript
import { GetAuthenticationSettings200ResponseSamlIdp } from './api';

const instance: GetAuthenticationSettings200ResponseSamlIdp = {
    entity_id,
    certificates,
    binding,
    mappings,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
