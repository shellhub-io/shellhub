# WebendpointTLS

Web endpoint TLS configuration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**enabled** | **boolean** | Whether TLS is enabled for this web endpoint | [default to undefined]
**verify** | **boolean** | Whether to verify the TLS certificate | [default to undefined]
**domain** | **string** | Domain for TLS verification | [default to undefined]

## Example

```typescript
import { WebendpointTLS } from './api';

const instance: WebendpointTLS = {
    enabled,
    verify,
    domain,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
