# GetCustomer200Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Customer\&#39;s ID. | [optional] [default to undefined]
**name** | **string** | Customer\&#39;s name. | [optional] [default to undefined]
**email** | **string** | Customer\&#39;s e-mail. | [optional] [default to undefined]
**payment_methods** | [**Array&lt;GetCustomer200ResponsePaymentMethodsInner&gt;**](GetCustomer200ResponsePaymentMethodsInner.md) | Customer\&#39;s payment methods. | [optional] [default to undefined]

## Example

```typescript
import { GetCustomer200Response } from './api';

const instance: GetCustomer200Response = {
    id,
    name,
    email,
    payment_methods,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
