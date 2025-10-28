# DeviceInfo

Device\'s info

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Device\&#39;s OS name | [optional] [default to undefined]
**pretty_name** | **string** | Device\&#39;s OS pretty name | [optional] [default to undefined]
**version** | **string** | Device\&#39;s OS version | [optional] [default to undefined]
**arch** | **string** | Device\&#39;s OS arch | [optional] [default to undefined]
**platform** | [**DeviceInfoPlatform**](DeviceInfoPlatform.md) |  | [optional] [default to undefined]

## Example

```typescript
import { DeviceInfo } from './api';

const instance: DeviceInfo = {
    id,
    pretty_name,
    version,
    arch,
    platform,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
