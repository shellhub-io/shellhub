# BillingApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**attachPaymentMethod**](#attachpaymentmethod) | **POST** /api/billing/paymentmethod/attach | Attach payment method|
|[**choiceDevices**](#choicedevices) | **POST** /api/billing/device-choice | Choice devices|
|[**createCustomer**](#createcustomer) | **POST** /api/billing/customer | Create customer|
|[**createSubscription**](#createsubscription) | **POST** /api/billing/subscription | Create subscription|
|[**detachPaymentMethod**](#detachpaymentmethod) | **POST** /api/billing/paymentmethod/detach | Detach payment method|
|[**evaluate**](#evaluate) | **POST** /api/billing/evaluate | Evaluate|
|[**getCustomer**](#getcustomer) | **GET** /api/billing/customer | Get Customer|
|[**getDevicesMostUsed**](#getdevicesmostused) | **GET** /api/billing/devices-most-used | Get devices most used|
|[**getSubscription**](#getsubscription) | **GET** /api/billing/subscription | Get subscription|
|[**report**](#report) | **POST** /api/billing/report | Report|
|[**setDefaultPaymentMethod**](#setdefaultpaymentmethod) | **POST** /api/billing/paymentmethod/default | Set default payment method|

# **attachPaymentMethod**
> attachPaymentMethod()

Attachs a payment method to a customer.

### Example

```typescript
import {
    BillingApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.attachPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to attach the payment method to a customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **choiceDevices**
> choiceDevices()

Choice devices when device\'s limit is rechead.

### Example

```typescript
import {
    BillingApi,
    Configuration,
    ChoiceDevicesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

let choiceDevicesRequest: ChoiceDevicesRequest; // (optional)

const { status, data } = await apiInstance.choiceDevices(
    choiceDevicesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **choiceDevicesRequest** | **ChoiceDevicesRequest**|  | |


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to choice devices. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createCustomer**
> createCustomer()

creates a new customer defining, optionaly, the default payment method.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.createCustomer();
```

### Parameters
This endpoint does not have any parameters.


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create a new customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**409** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createSubscription**
> createSubscription()

Create a subscription.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.createSubscription();
```

### Parameters
This endpoint does not have any parameters.


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create a new subscription. |  -  |
|**400** | Error |  -  |
|**402** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**409** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **detachPaymentMethod**
> detachPaymentMethod()

Detachs a payment method from a customer.

### Example

```typescript
import {
    BillingApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.detachPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to detach the payment method from a customer. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **evaluate**
> Evaluate200Response evaluate()

evaluate the namespace capabilities.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.evaluate();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Evaluate200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to evaluate the namespace. |  -  |
|**400** | Bad Request. |  -  |
|**401** | Unauthorized. |  -  |
|**403** | Forbidden. |  -  |
|**404** | Not found. |  -  |
|**424** | Failed dependency. |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getCustomer**
> GetCustomer200Response getCustomer()

Get the customer.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.getCustomer();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetCustomer200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a customer. |  -  |
|**400** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevicesMostUsed**
> Array<Device> getDevicesMostUsed()

Get the most used devices.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.getDevicesMostUsed();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**Array<Device>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the most used devices. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSubscription**
> GetSubscription200Response getSubscription()

Get the subscription.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

const { status, data } = await apiInstance.getSubscription();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetSubscription200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a subscription. |  -  |
|**402** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **report**
> report()

Report an action.

### Example

```typescript
import {
    BillingApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

let action: ReportActionParameter; // (default to undefined)

const { status, data } = await apiInstance.report(
    action
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **action** | **ReportActionParameter** |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to report a action. |  -  |
|**400** | Bad Request. |  -  |
|**401** | Unauthorized. |  -  |
|**402** | Payment Required. |  -  |
|**403** | Forbidden. |  -  |
|**404** | Not found. |  -  |
|**424** | Failed dependency. |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setDefaultPaymentMethod**
> setDefaultPaymentMethod()

Set default payment method to the customer.

### Example

```typescript
import {
    BillingApi,
    Configuration,
    AttachPaymentMethodRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new BillingApi(configuration);

let attachPaymentMethodRequest: AttachPaymentMethodRequest; // (optional)

const { status, data } = await apiInstance.setDefaultPaymentMethod(
    attachPaymentMethodRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **attachPaymentMethodRequest** | **AttachPaymentMethodRequest**|  | |


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to set default payment method. |  -  |
|**400** | Error |  -  |
|**401** | Error |  -  |
|**403** | Error |  -  |
|**404** | Error |  -  |
|**424** | Error |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

