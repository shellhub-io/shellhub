# GetSubscription200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Subscription\&#39;s ID. | [optional] [default to undefined]
**active** | **boolean** | Subscription\&#39;s active. | [optional] [default to undefined]
**status** | [**GetSubscription200ResponseStatus**](GetSubscription200ResponseStatus.md) |  | [optional] [default to undefined]
**end_at** | **number** | Subscription\&#39;s current period end. | [optional] [default to undefined]
**invoices** | [**Array&lt;GetSubscription200ResponseInvoicesInner&gt;**](GetSubscription200ResponseInvoicesInner.md) | Subscription\&#39;s invoices. | [optional] [default to undefined]

## Example

```typescript
import { GetSubscription200Response } from './api';

const instance: GetSubscription200Response = {
    id,
    active,
    status,
    end_at,
    invoices,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
