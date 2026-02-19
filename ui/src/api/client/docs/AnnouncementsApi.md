# AnnouncementsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**getAnnouncement**](#getannouncement) | **GET** /api/announcements/{uuid} | Get a announcement|
|[**listAnnouncements**](#listannouncements) | **GET** /api/announcements | List announcements|

# **getAnnouncement**
> Announcement getAnnouncement()

Get a announcement.

### Example

```typescript
import {
    AnnouncementsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AnnouncementsApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.getAnnouncement(
    uuid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uuid** | [**string**] |  | defaults to undefined|


### Return type

**Announcement**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a announcement. |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAnnouncements**
> Array<AnnouncementShort> listAnnouncements()

List the announcements posted by ShellHub Cloud.

### Example

```typescript
import {
    AnnouncementsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AnnouncementsApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ListAnnouncementsOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.listAnnouncements(
    page,
    perPage,
    orderBy
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **orderBy** | **ListAnnouncementsOrderByParameter** |  | (optional) defaults to undefined|


### Return type

**Array<AnnouncementShort>**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the announcements. |  * X-Total-Count -  <br>  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

