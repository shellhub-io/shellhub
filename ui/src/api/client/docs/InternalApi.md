# InternalApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**authSSHPublicKey**](#authsshpublickey) | **POST** /api/auth/ssh | Auth SSH public key|
|[**updateDeviceStatusOffline**](#updatedevicestatusoffline) | **POST** /internal/devices/{uid}/offline | Update device status to offline|

# **authSSHPublicKey**
> AuthSSHPublicKey200Response authSSHPublicKey()

Authenticate a SSH public key to ShellHub server.

### Example

```typescript
import {
    InternalApi,
    Configuration,
    AuthSSHPublicKeyRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new InternalApi(configuration);

let authSSHPublicKeyRequest: AuthSSHPublicKeyRequest; // (optional)

const { status, data } = await apiInstance.authSSHPublicKey(
    authSSHPublicKeyRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **authSSHPublicKeyRequest** | **AuthSSHPublicKeyRequest**|  | |


### Return type

**AuthSSHPublicKey200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to auth a SSH public key. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceStatusOffline**
> updateDeviceStatusOffline()

Update device\'s status to offiline.

### Example

```typescript
import {
    InternalApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new InternalApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.updateDeviceStatusOffline(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to update device status to offline. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

