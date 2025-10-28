# StatsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getStats**](#getstats) | **GET** /admin/api/stats | Get stats|

# **getStats**
> GetStats200Response getStats()

Get stats about the ShellHub instance.

### Example

```typescript
import {
    StatsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new StatsApi(configuration);

const { status, data } = await apiInstance.getStats();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetStats200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get stats about the ShellHub instance |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

