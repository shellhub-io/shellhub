# GetCustomer200ResponsePaymentMethodsInner


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**id** | **string** | Payment method\&#39;s ID. | [optional] [default to undefined]
**number** | **string** | Payment method card\&#39;s number. | [optional] [default to undefined]
**brand** | **string** | Payment method card\&#39;s brand. | [optional] [default to undefined]
**exp_month** | **number** | Payment method card\&#39;s expiration month. | [optional] [default to undefined]
**exp_year** | **number** | Payment method card\&#39;s expiration year. | [optional] [default to undefined]
**cvc** | **string** | Payment method card\&#39;s CVC. | [optional] [default to undefined]
**_default** | **boolean** | Payment method default status. | [optional] [default to undefined]

## Example

```typescript
import { GetCustomer200ResponsePaymentMethodsInner } from './api';

const instance: GetCustomer200ResponsePaymentMethodsInner = {
    id,
    number,
    brand,
    exp_month,
    exp_year,
    cvc,
    _default,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
