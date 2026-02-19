# ContainersApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**deleteContainer**](#deletecontainer) | **DELETE** /api/containers/{uid} | Delete container|
|[**getContainer**](#getcontainer) | **GET** /api/containers/{uid} | Get container|
|[**getContainers**](#getcontainers) | **GET** /api/containers | Get containers|
|[**updateContainer**](#updatecontainer) | **PUT** /api/containers/{uid} | Update container|
|[**updateContainerStatus**](#updatecontainerstatus) | **PATCH** /api/containers/{uid}/{status} | Update container status|

# **deleteContainer**
> deleteContainer()

Delete a container.

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.deleteContainer(
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
|**200** | Success to delete a container. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainer**
> Device getContainer()

Get a container.

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.getContainer(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


### Return type

**Device**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success get a container. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getContainers**
> Array<Device> getContainers()

Get a list of containers.

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let status: DeviceStatus; //Container\'s status (optional) (default to undefined)
let sortBy: string; //Container\'s property to sort of (optional) (default to 'last_seen')
let orderBy: GetContainersOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.getContainers(
    filter,
    page,
    perPage,
    status,
    sortBy,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **status** | **DeviceStatus** | Container\&#39;s status | (optional) defaults to undefined|
| **sortBy** | [**string**] | Container\&#39;s property to sort of | (optional) defaults to 'last_seen'|
| **orderBy** | **GetContainersOrderByParameter** |  | (optional) defaults to undefined|


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
|**200** | Success to get a list of containers. |  * X-Total-Count -  <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateContainer**
> updateContainer()

Update container\'s data.

### Example

```typescript
import {
    ContainersApi,
    Configuration,
    UpdateContainerRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let updateContainerRequest: UpdateContainerRequest; // (optional)

const { status, data } = await apiInstance.updateContainer(
    uid,
    updateContainerRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateContainerRequest** | **UpdateContainerRequest**|  | |
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|


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
|**200** | Success to update container\&#39;s data. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateContainerStatus**
> updateContainerStatus()

Update container\'s status.

### Example

```typescript
import {
    ContainersApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new ContainersApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let status: UpdateDeviceStatusStatusParameter; //Container\'s status (default to undefined)

const { status, data } = await apiInstance.updateContainerStatus(
    uid,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **status** | **UpdateDeviceStatusStatusParameter** | Container\&#39;s status | defaults to undefined|


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
|**200** | Success to update container status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

