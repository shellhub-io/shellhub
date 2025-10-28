# LicenseApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getLicense**](#getlicense) | **GET** /admin/api/license | Get license data|
|[**sendLicense**](#sendlicense) | **POST** /admin/api/license | Send license data|

# **getLicense**
> GetLicense200Response getLicense()

Get the license data.

### Example

```typescript
import {
    LicenseApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new LicenseApi(configuration);

const { status, data } = await apiInstance.getLicense();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetLicense200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get data. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **sendLicense**
> sendLicense()

Send license data

### Example

```typescript
import {
    LicenseApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new LicenseApi(configuration);

let file: File; //License\\\'s file (optional) (default to undefined)

const { status, data } = await apiInstance.sendLicense(
    file
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **file** | [**File**] | License\\\&#39;s file | (optional) defaults to undefined|


### Return type

void (empty response body)

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to send license data |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

