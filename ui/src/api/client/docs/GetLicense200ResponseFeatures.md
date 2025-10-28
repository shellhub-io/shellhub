# GetLicense200ResponseFeatures

License\'s features

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**devices** | **number** | Number of devices allowed   &#x60;-1&#x60; means unlimited number of devices and any other number means the number of devices allowed  | [default to undefined]
**session_recording** | **boolean** | Session recording status | [default to undefined]
**firewall_rules** | **boolean** | Firewall rules status | [default to undefined]
**reports** | **boolean** | Reports status | [default to undefined]
**login_link** | **boolean** | Login link status | [default to undefined]
**billing** | **boolean** | Billing status | [default to undefined]

## Example

```typescript
import { GetLicense200ResponseFeatures } from './api';

const instance: GetLicense200ResponseFeatures = {
    devices,
    session_recording,
    firewall_rules,
    reports,
    login_link,
    billing,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
