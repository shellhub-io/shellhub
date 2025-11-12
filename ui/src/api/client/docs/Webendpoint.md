# Webendpoint


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**address** | **string** | Web endpoint\&#39;s unique address | [optional] [default to undefined]
**full_address** | **string** | Full webendpoint address including domain | [optional] [default to undefined]
**namespace** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**device_uid** | **string** | Device\&#39;s UID | [optional] [default to undefined]
**device** | [**Device**](Device.md) |  | [optional] [default to undefined]
**host** | **string** | Web endpoint\&#39;s agent host address | [optional] [default to undefined]
**port** | **number** | Web endpoint\&#39;s agent port number | [optional] [default to undefined]
**ttl** | **number** | Web endpoint\&#39;s time to live in seconds | [optional] [default to undefined]
**tls** | [**WebendpointTLS**](WebendpointTLS.md) |  | [optional] [default to undefined]
**expires_in** | **string** | Web endpoint\&#39;s expiration date | [optional] [default to undefined]
**created_at** | **string** | Web endpoint\&#39;s creation date | [optional] [default to undefined]

## Example

```typescript
import { Webendpoint } from './api';

const instance: Webendpoint = {
    address,
    full_address,
    namespace,
    device_uid,
    device,
    host,
    port,
    ttl,
    tls,
    expires_in,
    created_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
