# ApiKeysApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**apiKeyCreate**](#apikeycreate) | **POST** /api/namespaces/api-key | Creates an API key.|
|[**apiKeyDelete**](#apikeydelete) | **DELETE** /api/namespaces/api-key/{key} | Delete an API key|
|[**apiKeyList**](#apikeylist) | **GET** /api/namespaces/api-key | List API Keys|
|[**apiKeyUpdate**](#apikeyupdate) | **PATCH** /api/namespaces/api-key/{key} | Update an API key|

# **apiKeyCreate**
> ApiKeyWithID apiKeyCreate()

The `created_by`, `tenant_id`, and `role` (unless provided in the request body) values will be obtained from the JWT token. 

### Example

```typescript
import {
    ApiKeysApi,
    Configuration,
    ApiKeyCreate
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let apiKeyCreate: ApiKeyCreate; // (optional)

const { status, data } = await apiInstance.apiKeyCreate(
    apiKeyCreate
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **apiKeyCreate** | **ApiKeyCreate**|  | |


### Return type

**ApiKeyWithID**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyDelete**
> apiKeyDelete()


### Example

```typescript
import {
    ApiKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let key: string; // (default to undefined)

const { status, data } = await apiInstance.apiKeyDelete(
    key
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **key** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyList**
> Array<ApiKey> apiKeyList()


### Example

```typescript
import {
    ApiKeysApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ApiKeyListOrderByParameter; // (optional) (default to undefined)
let sortBy: string; //The property to sort of. (optional) (default to 'expires_in')

const { status, data } = await apiInstance.apiKeyList(
    page,
    perPage,
    orderBy,
    sortBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **orderBy** | **ApiKeyListOrderByParameter** |  | (optional) defaults to undefined|
| **sortBy** | [**string**] | The property to sort of. | (optional) defaults to 'expires_in'|


### Return type

**Array<ApiKey>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success. |  * X-Total-Count -  <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiKeyUpdate**
> apiKeyUpdate()


### Example

```typescript
import {
    ApiKeysApi,
    Configuration,
    ApiKeyUpdate
} from './api';

const configuration = new Configuration();
const apiInstance = new ApiKeysApi(configuration);

let key: string; // (default to undefined)
let apiKeyUpdate: ApiKeyUpdate; // (optional)

const { status, data } = await apiInstance.apiKeyUpdate(
    key,
    apiKeyUpdate
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **apiKeyUpdate** | **ApiKeyUpdate**|  | |
| **key** | [**string**] |  | defaults to undefined|


### Return type

void (empty response body)

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

