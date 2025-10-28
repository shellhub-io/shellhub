# Session


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**uid** | **string** | Session\&#39;s UID | [optional] [default to undefined]
**device_uid** | **string** | Device\&#39;s UID | [optional] [default to undefined]
**device** | [**Device**](Device.md) |  | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**username** | **string** | Session\&#39;s username | [optional] [default to undefined]
**ip_address** | **string** | Session\&#39;s IP address | [optional] [default to undefined]
**started_at** | **string** | Session\&#39;s started date | [optional] [default to undefined]
**last_seen** | **string** | Session\&#39;s last seen date | [optional] [default to undefined]
**active** | **boolean** | Session\&#39;s active status | [optional] [default to undefined]
**authenticated** | **boolean** | Session\&#39;s authenticated status | [optional] [default to undefined]
**recorded** | **boolean** | Session\&#39;s recorded status | [optional] [default to undefined]
**type** | [**SessionType**](SessionType.md) |  | [optional] [default to undefined]
**term** | **string** | Session\&#39;s terminal | [optional] [default to undefined]
**position** | [**SessionPosition**](SessionPosition.md) |  | [optional] [default to undefined]
**events** | [**SessionEvents**](SessionEvents.md) |  | [optional] [default to undefined]

## Example

```typescript
import { Session } from './api';

const instance: Session = {
    uid,
    device_uid,
    device,
    tenant_id,
    username,
    ip_address,
    started_at,
    last_seen,
    active,
    authenticated,
    recorded,
    type,
    term,
    position,
    events,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
