# WebEndpointsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createWebEndpoint**](#createwebendpoint) | **POST** /api/web-endpoints | Create a web-endpoint|
|[**deleteWebEndpoint**](#deletewebendpoint) | **DELETE** /api/web-endpoints/{address} | Delete a web-endpoint|
|[**listWebEndpoints**](#listwebendpoints) | **GET** /api/web-endpoints | List web-endpoints|

# **createWebEndpoint**
> Webendpoint createWebEndpoint(createWebEndpointRequest)

Creates a new web-endpoint for a device.

### Example

```typescript
import {
    WebEndpointsApi,
    Configuration,
    CreateWebEndpointRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new WebEndpointsApi(configuration);

let createWebEndpointRequest: CreateWebEndpointRequest; //

const { status, data } = await apiInstance.createWebEndpoint(
    createWebEndpointRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createWebEndpointRequest** | **CreateWebEndpointRequest**|  | |


### Return type

**Webendpoint**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Web-endpoint created successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteWebEndpoint**
> deleteWebEndpoint()

Deletes a web-endpoint by address.

### Example

```typescript
import {
    WebEndpointsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new WebEndpointsApi(configuration);

let address: string; //Tunnel\'s address (default to undefined)

const { status, data } = await apiInstance.deleteWebEndpoint(
    address
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **address** | [**string**] | Tunnel\&#39;s address | defaults to undefined|


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
|**200** | Web-endpoint deleted successfully. |  -  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listWebEndpoints**
> Array<Webendpoint> listWebEndpoints()

List all web-endpoints in the namespace.

### Example

```typescript
import {
    WebEndpointsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new WebEndpointsApi(configuration);

let filter: string; //Web endpoint\'s filter  Filter field receives a base64 enconded JSON object for limit a search.  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let sortBy: ListWebEndpointsSortByParameter; //Field to sort by (optional) (default to undefined)
let orderBy: ListWebEndpointsOrderByParameter; //Sort order (asc or desc) (optional) (default to undefined)

const { status, data } = await apiInstance.listWebEndpoints(
    filter,
    page,
    perPage,
    sortBy,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Web endpoint\&#39;s filter  Filter field receives a base64 enconded JSON object for limit a search.  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **sortBy** | **ListWebEndpointsSortByParameter** | Field to sort by | (optional) defaults to undefined|
| **orderBy** | **ListWebEndpointsOrderByParameter** | Sort order (asc or desc) | (optional) defaults to undefined|


### Return type

**Array<Webendpoint>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the web-endpoints. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

