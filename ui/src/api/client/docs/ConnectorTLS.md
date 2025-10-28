# ConnectorTLS


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ca** | **string** | Certificate Authority used to generate the Cert for the server and the client. | [default to undefined]
**cert** | **string** | Certificate generated from the CA certificate and used by the client to authorize the connection to the Docker Engine. | [default to undefined]
**key** | **string** | Private key for the certificate on the Cert field. | [default to undefined]

## Example

```typescript
import { ConnectorTLS } from './api';

const instance: ConnectorTLS = {
    ca,
    cert,
    key,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
