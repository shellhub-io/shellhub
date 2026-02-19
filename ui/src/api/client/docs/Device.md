# Device


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**uid** | **string** | Device\&#39;s UID | [default to undefined]
**name** | **string** | Device\&#39;s name   By default, the name is the device\&#39;s MAC address when it just added.  | [default to undefined]
**identity** | [**DeviceIdentity**](DeviceIdentity.md) |  | [optional] [default to undefined]
**info** | [**DeviceInfo**](DeviceInfo.md) |  | [optional] [default to undefined]
**public_key** | **string** | Device\&#39;s public key. | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [default to undefined]
**last_seen** | **string** | Device\&#39;s last seen date | [default to undefined]
**online** | **boolean** | Device\&#39;s availability status | [default to undefined]
**namespace** | **string** | Namespace\&#39;s name | [optional] [default to undefined]
**status** | [**DeviceStatus**](DeviceStatus.md) |  | [default to undefined]
**status_update_at** | **string** | Device\&#39;s status update date | [optional] [default to undefined]
**created_at** | **string** | Device\&#39;s creation date | [default to undefined]
**remote_addr** | **string** | Device\&#39;s remote address | [optional] [default to undefined]
**position** | [**DevicePosition**](DevicePosition.md) |  | [optional] [default to undefined]
**tags** | [**Array&lt;Tag&gt;**](Tag.md) | Device\&#39;s Tags list | [optional] [default to undefined]
**public_url** | **boolean** | Device\&#39;s public URL status. | [optional] [default to undefined]
**acceptable** | **boolean** | Device\&#39;s acceptable  The value \&quot;acceptable\&quot; is based on the number of devices removed and already accepted into a namespace. All devices are \&quot;acceptable\&quot; unless the \&quot;namespace.max_devices\&quot; is reached. This limit is set based on the sum up of accepted and removed devices into the namespace. When this limit is reached, only removed devices between 720 hours or 30 days are set to \&quot;acceptable\&quot;.  | [optional] [default to undefined]

## Example

```typescript
import { Device } from './api';

const instance: Device = {
    uid,
    name,
    identity,
    info,
    public_key,
    tenant_id,
    last_seen,
    online,
    namespace,
    status,
    status_update_at,
    created_at,
    remote_addr,
    position,
    tags,
    public_url,
    acceptable,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
