# GetLicense200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | License\&#39;s ID | [default to undefined]
**expired** | **boolean** | License\&#39;s expired status | [default to undefined]
**about_to_expire** | **boolean** | License\&#39;s about to expire status | [default to undefined]
**grace_period** | **boolean** | License\&#39;s grace period status | [default to undefined]
**issued_at** | **number** | License\&#39;s issued at  It is the number of seconds elapsed since January 1, 1970 UTC. | [default to undefined]
**starts_at** | **number** | License\&#39;s started at  It is the number of seconds elapsed since January 1, 1970 UTC. | [default to undefined]
**expires_at** | **number** | License\&#39;s expired at  It is the number of seconds elapsed since January 1, 1970 UTC.  &#x60;-1&#x60; means license does not expire | [default to undefined]
**allowed_regions** | [**Array&lt;GetLicense200ResponseAllowedRegionsInner&gt;**](GetLicense200ResponseAllowedRegionsInner.md) | License\&#39;s allowed regions  It is a list of regions in &#x60;ISO 3166-1 alpha-2&#x60; format. | [default to undefined]
**customer** | [**GetLicense200ResponseCustomer**](GetLicense200ResponseCustomer.md) |  | [default to undefined]
**features** | [**GetLicense200ResponseFeatures**](GetLicense200ResponseFeatures.md) |  | [default to undefined]

## Example

```typescript
import { GetLicense200Response } from './api';

const instance: GetLicense200Response = {
    id,
    expired,
    about_to_expire,
    grace_period,
    issued_at,
    starts_at,
    expires_at,
    allowed_regions,
    customer,
    features,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
