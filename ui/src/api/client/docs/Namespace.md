# Namespace


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | Namespace\&#39;s name | [optional] [default to undefined]
**owner** | **string** | User\&#39;s ID. | [optional] [default to undefined]
**tenant_id** | **string** | Namespace\&#39;s tenant ID | [optional] [default to undefined]
**members** | [**Array&lt;NamespaceMembersInner&gt;**](NamespaceMembersInner.md) | Namespace\&#39;s members | [optional] [default to undefined]
**settings** | [**NamespaceSettings**](NamespaceSettings.md) |  | [optional] [default to undefined]
**max_devices** | **number** | Namespace\&#39;s max device numbers | [optional] [default to 3]
**device_count** | **number** | Namespace\&#39;s total devices | [optional] [default to undefined]
**created_at** | **string** | Namespace\&#39;s creation date | [optional] [default to undefined]
**billing** | **object** | Namespace\&#39;s billing | [optional] [default to undefined]
**devices_pending_count** | **number** | Number of devices currently in pending status awaiting approval | [optional] [default to undefined]
**devices_accepted_count** | **number** | Number of devices that have been accepted and are active in the namespace | [optional] [default to undefined]
**devices_rejected_count** | **number** | Number of devices that have been explicitly rejected from the namespace | [optional] [default to undefined]

## Example

```typescript
import { Namespace } from './api';

const instance: Namespace = {
    name,
    owner,
    tenant_id,
    members,
    settings,
    max_devices,
    device_count,
    created_at,
    billing,
    devices_pending_count,
    devices_accepted_count,
    devices_rejected_count,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
