# AdminApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**adminDeleteUser**](#admindeleteuser) | **DELETE** /admin/api/users/{id} | Delete user|
|[**adminResetUserPassword**](#adminresetuserpassword) | **PATCH** /admin/api/users/{id}/password/reset | Reset user password|
|[**adminUpdateUser**](#adminupdateuser) | **PUT** /admin/api/users/{id} | Update user|
|[**configureLocalAuthentication**](#configurelocalauthentication) | **PUT** /admin/api/authentication/local | Configure Local Authentication|
|[**configureSAMLAuthentication**](#configuresamlauthentication) | **PUT** /admin/api/authentication/saml | Configure SAML Authentication|
|[**createAnnouncement**](#createannouncement) | **POST** /admin/api/announcements | Create an announcement|
|[**createFirewallRuleAdmin**](#createfirewallruleadmin) | **POST** /admin/api/firewall/rules | Create firewall rule Admin|
|[**createNamespaceAdmin**](#createnamespaceadmin) | **POST** /admin/api/namespaces/{tenant} | Create namespace admin|
|[**createUserAdmin**](#createuseradmin) | **POST** /admin/api/users | Create a User admin|
|[**deleteAnnouncement**](#deleteannouncement) | **DELETE** /admin/api/announcements/{uuid} | Delete an announcement|
|[**deleteDeviceAdmin**](#deletedeviceadmin) | **DELETE** /admin/api/devices/{uid} | Delete device admin|
|[**deleteFirewallRuleAdmin**](#deletefirewallruleadmin) | **DELETE** /admin/api/firewall/rules/{id} | Delete firewall rule admin|
|[**deleteNamespaceAdmin**](#deletenamespaceadmin) | **DELETE** /admin/api/namespaces/{tenant} | Delete namespace admin|
|[**editNamespaceAdmin**](#editnamespaceadmin) | **PUT** /admin/api/namespaces-update/{tenantID} | Edit namespace admin|
|[**exportNamespaces**](#exportnamespaces) | **GET** /admin/api/export/namespaces | export namespace|
|[**exportUsers**](#exportusers) | **GET** /admin/api/export/users | export users|
|[**getAnnouncementAdmin**](#getannouncementadmin) | **GET** /admin/api/announcements/{uuid} | Get a announcement|
|[**getAuthenticationSettings**](#getauthenticationsettings) | **GET** /admin/api/authentication | Get Authentication Settings|
|[**getDeviceAdmin**](#getdeviceadmin) | **GET** /admin/api/devices/{uid} | Get device admin|
|[**getDevicesAdmin**](#getdevicesadmin) | **GET** /admin/api/devices | Get devices admin|
|[**getFirewallRuleAdmin**](#getfirewallruleadmin) | **GET** /admin/api/firewall/rules/{id} | Get firewall rule admin|
|[**getFirewallRulesAdmin**](#getfirewallrulesadmin) | **GET** /admin/api/firewall/rules | Get firewall rules Admin|
|[**getLicense**](#getlicense) | **GET** /admin/api/license | Get license data|
|[**getNamespaceAdmin**](#getnamespaceadmin) | **GET** /admin/api/namespaces/{tenant} | Get namespace admin|
|[**getNamespacesAdmin**](#getnamespacesadmin) | **GET** /admin/api/namespaces | Get namespaces admin|
|[**getSessionAdmin**](#getsessionadmin) | **GET** /admin/api/sessions/{uid} | Get session admin|
|[**getSessionsAdmin**](#getsessionsadmin) | **GET** /admin/api/sessions | Get sessions admin|
|[**getStats**](#getstats) | **GET** /admin/api/stats | Get stats|
|[**getUser**](#getuser) | **GET** /admin/api/users/{id} | Get user|
|[**getUserTokenAdmin**](#getusertokenadmin) | **GET** /admin/api/auth/token/{id} | Get user token|
|[**getUsers**](#getusers) | **GET** /admin/api/users | Get users|
|[**listAnnouncementsAdmin**](#listannouncementsadmin) | **GET** /admin/api/announcements | List announcements|
|[**loginAdmin**](#loginadmin) | **POST** /admin/api/login | Login on Admin|
|[**sendLicense**](#sendlicense) | **POST** /admin/api/license | Send license data|
|[**setSessionAuthenticationStatusAdmin**](#setsessionauthenticationstatusadmin) | **POST** /admin/api/sessions/{uid} | Set session authentication status admin|
|[**updateAnnouncement**](#updateannouncement) | **PUT** /admin/api/announcements/{uuid} | Update an announcement|
|[**updateDeviceNameAdmin**](#updatedevicenameadmin) | **PATCH** /admin/api/devices/{uid} | Update device name Admin|
|[**updateDeviceStatusAdmin**](#updatedevicestatusadmin) | **PATCH** /admin/api/devices/{uid}/{status} | Update status Admin|
|[**updateFirewallRuleAdmin**](#updatefirewallruleadmin) | **PUT** /admin/api/firewall/rules/{id} | Update firewall rule admin|

# **adminDeleteUser**
> adminDeleteUser()

Delete a user.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.adminDeleteUser(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminResetUserPassword**
> AdminResetUserPassword200Response adminResetUserPassword()

Resets the password for a specified user. It\'s particularly useful for users who initially authenticated via SAML and therefore may not have a password set up. This enables them to maintain access even if their original authentication method becomes unavailable (e.g., if SAML authentication is disabled).  The endpoint generates a secure 16-character random password that includes: - Uppercase letters - Lowercase letters - Numbers - Special characters  Users are strongly encouraged to change this temporary password after their first successful authentication. If the user already has a password, a `400 Bad Request` status code will be returned. 

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; //The ID of the user whose password needs to be reset (default to undefined)

const { status, data } = await apiInstance.adminResetUserPassword(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | The ID of the user whose password needs to be reset | defaults to undefined|


### Return type

**AdminResetUserPassword200Response**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Password successfully reset |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminUpdateUser**
> adminUpdateUser()

Update a user.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    UserAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)
let userAdminRequest: UserAdminRequest; // (optional)

const { status, data } = await apiInstance.adminUpdateUser(
    id,
    userAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **userAdminRequest** | **UserAdminRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **configureLocalAuthentication**
> configureLocalAuthentication(configureLocalAuthenticationRequest)

Configure local authentication settings for the ShellHub instance.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    ConfigureLocalAuthenticationRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let configureLocalAuthenticationRequest: ConfigureLocalAuthenticationRequest; //

const { status, data } = await apiInstance.configureLocalAuthentication(
    configureLocalAuthenticationRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **configureLocalAuthenticationRequest** | **ConfigureLocalAuthenticationRequest**|  | |


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
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **configureSAMLAuthentication**
> configureSAMLAuthentication(configureSAMLAuthenticationRequest)

Configure SAML authentication settings for the ShellHub instance.  The NameID in the SAML assertion from the IdP must be configured to use a format that provides a unique and persistent identifier for each user. This could be a persistent ID, email address, or any other attribute that uniquely identifies the user within your IdP. 

### Example

```typescript
import {
    AdminApi,
    Configuration,
    ConfigureSAMLAuthenticationRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let configureSAMLAuthenticationRequest: ConfigureSAMLAuthenticationRequest; //

const { status, data } = await apiInstance.configureSAMLAuthentication(
    configureSAMLAuthenticationRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **configureSAMLAuthenticationRequest** | **ConfigureSAMLAuthenticationRequest**|  | |


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
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createAnnouncement**
> Announcement createAnnouncement()

Create an announcement.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    CreateAnnouncementRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let createAnnouncementRequest: CreateAnnouncementRequest; // (optional)

const { status, data } = await apiInstance.createAnnouncement(
    createAnnouncementRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAnnouncementRequest** | **CreateAnnouncementRequest**|  | |


### Return type

**Announcement**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createFirewallRuleAdmin**
> FirewallRulesResponse createFirewallRuleAdmin()

Create a firewall rule.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.createFirewallRuleAdmin(
    firewallRulesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **firewallRulesRequest** | **FirewallRulesRequest**|  | |


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create firewall rule. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createNamespaceAdmin**
> Namespace createNamespaceAdmin()

Create a namespace.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    CreateNamespaceRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)
let createNamespaceRequest: CreateNamespaceRequest; // (optional)

const { status, data } = await apiInstance.createNamespaceAdmin(
    tenant,
    createNamespaceRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createNamespaceRequest** | **CreateNamespaceRequest**|  | |
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to create a namespace. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createUserAdmin**
> createUserAdmin()

Create a User.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    UserAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let userAdminRequest: UserAdminRequest; // (optional)

const { status, data } = await apiInstance.createUserAdmin(
    userAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **userAdminRequest** | **UserAdminRequest**|  | |


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
|**200** | Success |  -  |
|**400** | Invalid Fields |  -  |
|**401** | Unauthorized |  -  |
|**409** | Conflict Fields |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteAnnouncement**
> Announcement deleteAnnouncement()

Delete an announcement.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.deleteAnnouncement(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to delete an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteDeviceAdmin**
> deleteDeviceAdmin()

Delete a device.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.deleteDeviceAdmin(
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
|**200** | Success to delete a device. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteFirewallRuleAdmin**
> deleteFirewallRuleAdmin()

Delete a firewall rule.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteFirewallRuleAdmin(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


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
|**200** | Success to delete a firewall rule. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteNamespaceAdmin**
> deleteNamespaceAdmin()

Delete a namespace.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.deleteNamespaceAdmin(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


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
|**200** | Success to delete a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **editNamespaceAdmin**
> editNamespaceAdmin()

Edit a namespace.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    Namespace
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let tenantID: string; //Namespace\'s tenant ID (default to undefined)
let namespace: Namespace; // (optional)

const { status, data } = await apiInstance.editNamespaceAdmin(
    tenantID,
    namespace
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **namespace** | **Namespace**|  | |
| **tenantID** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


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
|**200** | Success to edit a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **exportNamespaces**
> File exportNamespaces()

Export namespaces to csv file. This endpoint has been deprecated and will be removed in v1.0.0. 

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let filter: string; //Namespace\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `devices` where the value should be \'gt\' `0`.  An example of JSON object will looks like this:  ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"devices\",           \"operator\":\"eq\",           \"value\":\"0\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJkZXZpY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0=`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.exportNamespaces(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Namespace\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;devices&#x60; where the value should be \&#39;gt\&#39; &#x60;0&#x60;.  An example of JSON object will looks like this:  &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;devices\&quot;,           \&quot;operator\&quot;:\&quot;eq\&quot;,           \&quot;value\&quot;:\&quot;0\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJkZXZpY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**File**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to export namespaces. |  -  |
|**204** | No content. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **exportUsers**
> File exportUsers()

Export users to csv file.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let filter: string; //User\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `namespaces` where the value should be `eq` to `0`.  An example of JSON object will looks like this:  ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"namespace\",           \"operator\":\"eq\",           \"value\":\"0\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lc3BhY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0=`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.exportUsers(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | User\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;namespaces&#x60; where the value should be &#x60;eq&#x60; to &#x60;0&#x60;.  An example of JSON object will looks like this:  &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;namespace\&quot;,           \&quot;operator\&quot;:\&quot;eq\&quot;,           \&quot;value\&quot;:\&quot;0\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lc3BhY2VzIiwib3BlcmF0b3IiOiJndCIsInZhbHVlIjoiMCJ9fV0&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**File**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/octet-stream, application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to export users. |  -  |
|**204** | No content. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAnnouncementAdmin**
> Announcement getAnnouncementAdmin()

Get a announcement.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uuid: string; // (default to undefined)

const { status, data } = await apiInstance.getAnnouncementAdmin(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

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

# **getAuthenticationSettings**
> GetAuthenticationSettings200Response getAuthenticationSettings()

Retrieves the current authentication settings.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

const { status, data } = await apiInstance.getAuthenticationSettings();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**GetAuthenticationSettings200Response**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully retrieved the authentication settings. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDeviceAdmin**
> Device getDeviceAdmin()

Get a device.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; //Device\'s UID (default to undefined)

const { status, data } = await apiInstance.getDeviceAdmin(
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
|**200** | Success to get a device. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getDevicesAdmin**
> Array<Device> getDevicesAdmin()

Get a list of devices.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let filter: string; //Device\'s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `linux`.  If you want get only Devices name as `Linux`, the JSON object will looks like this   ```json   [     {       \"type\":\"property\",       \"params\":         {           \"name\":\"name\",           \"operator\":\"contains\",           \"value\":\"linux\"         }     }   ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let status: DeviceStatus; //Device\'s status (optional) (default to undefined)
let sortBy: string; //Device\'s property to sort of (optional) (default to undefined)
let orderBy: GetDevicesAdminOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.getDevicesAdmin(
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
| **filter** | [**string**] | Device\&#39;s filter   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;linux&#x60;.  If you want get only Devices name as &#x60;Linux&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json   [     {       \&quot;type\&quot;:\&quot;property\&quot;,       \&quot;params\&quot;:         {           \&quot;name\&quot;:\&quot;name\&quot;,           \&quot;operator\&quot;:\&quot;contains\&quot;,           \&quot;value\&quot;:\&quot;linux\&quot;         }     }   ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|
| **status** | **DeviceStatus** | Device\&#39;s status | (optional) defaults to undefined|
| **sortBy** | [**string**] | Device\&#39;s property to sort of | (optional) defaults to undefined|
| **orderBy** | **GetDevicesAdminOrderByParameter** |  | (optional) defaults to undefined|


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
|**200** | Success to get a list of devices. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFirewallRuleAdmin**
> FirewallRulesResponse getFirewallRuleAdmin()

Get a firewall rule.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getFirewallRuleAdmin(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get firewall rule. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFirewallRulesAdmin**
> Array<FirewallRulesResponse> getFirewallRulesAdmin()

Get a list of firewall rules.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getFirewallRulesAdmin(
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<FirewallRulesResponse>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get firewall rules. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getLicense**
> GetLicense200Response getLicense()

Get the license data.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

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

# **getNamespaceAdmin**
> Namespace getNamespaceAdmin()

Get a namespace.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let tenant: string; //Namespace\'s tenant ID (default to undefined)

const { status, data } = await apiInstance.getNamespaceAdmin(
    tenant
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **tenant** | [**string**] | Namespace\&#39;s tenant ID | defaults to undefined|


### Return type

**Namespace**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a namespace. |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getNamespacesAdmin**
> Array<Namespace> getNamespacesAdmin()

Returns a list of namespaces.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let filter: string; //Namespaces\'s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called `type`, it will filter by a `property` called `name` where the value should `contains` `examplespace`.  If you want get only Namespaces name as `examplespace`, the JSON object will looks like this   ```json [   {     \"type\":\"property\",     \"params\":{       \"name\":\"name\",       \"operator\":\"contains\",       \"value\":\"examplespace\"     }   } ] ```  So, the output encoded string will result on: `W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ==`  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getNamespacesAdmin(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Namespaces\&#39;s filter.   Filter field receives a base64 enconded JSON object for limit a search. The JSON object should have a property called &#x60;type&#x60;, it will filter by a &#x60;property&#x60; called &#x60;name&#x60; where the value should &#x60;contains&#x60; &#x60;examplespace&#x60;.  If you want get only Namespaces name as &#x60;examplespace&#x60;, the JSON object will looks like this   &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;:\&quot;property\&quot;,     \&quot;params\&quot;:{       \&quot;name\&quot;:\&quot;name\&quot;,       \&quot;operator\&quot;:\&quot;contains\&quot;,       \&quot;value\&quot;:\&quot;examplespace\&quot;     }   } ] &#x60;&#x60;&#x60;  So, the output encoded string will result on: &#x60;W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZXhhbXBsZXNwYWNlIn19XQ&#x3D;&#x3D;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Namespace>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a namespace list. |  * X-Total-Count - Namespaces\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessionAdmin**
> Session getSessionAdmin()

Get a session.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; // (default to undefined)

const { status, data } = await apiInstance.getSessionAdmin(
    uid
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] |  | defaults to undefined|


### Return type

**Session**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a session. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getSessionsAdmin**
> Array<Session> getSessionsAdmin()

Get a list sessions.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getSessionsAdmin(
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<Session>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get list of sessions. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getStats**
> GetStats200Response getStats()

Get stats about the ShellHub instance.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

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

# **getUser**
> GetUser200Response getUser()

Get a user.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getUser(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] |  | defaults to undefined|


### Return type

**GetUser200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a user. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserTokenAdmin**
> GetUserTokenAdmin200Response getUserTokenAdmin()

Get user JWT token to login.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; //User\'s ID (default to undefined)

const { status, data } = await apiInstance.getUserTokenAdmin(
    id
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **id** | [**string**] | User\&#39;s ID | defaults to undefined|


### Return type

**GetUserTokenAdmin200Response**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the JWT token to login. |  -  |
|**401** | Unauthorized |  -  |
|**404** | Not Found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUsers**
> Array<UserAdminResponse> getUsers()

Get a list of users.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let filter: string; //Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: ```typescript interface ParamProperty {   name: string;   operator: \"contains\" | \"eq\" | \"bool\" | \"gt\" | \"lt\";   value: string; }  interface ParamOperator {   name: \"and\" | \"or\"; }  interface Filter {   type: \"property\" | \"operator\";   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array<Filter>; }  ```  ## Examples  This is a example to filter and get only the resource what property \"confirmed\" is \"true\" ```json [   {   \"type\": \"property\",   \"params\": {       \"name\": \"confirmed\",       \"operator\": \"bool\",       \"value\": \"true\"       }   } ] ```  This one, filter resource by the property \"id\" inside \"info\" structure when it is equal to \"manjaro\" and online property is set to \"true\" ```json [   {     \"type\": \"property\",     \"params\": {       \"name\": \"info.id\",       \"operator\": \"eq\",       \"value\": \"manjaro\"     }   },   {     \"type\": \"property\",     \"params\": {       \"name\": \"online\",       \"operator\": \"bool\",       \"value\": \"true\"     }   },   {     \"type\": \"operator\",     \"params\": {       \"name\": \"and\"     }   } ] ```  (optional) (default to undefined)
let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getUsers(
    filter,
    page,
    perPage
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **filter** | [**string**] | Filter field receives a JSON object enconded as base64 string for limit a search.  The JSON enconded must follow these interafaces: &#x60;&#x60;&#x60;typescript interface ParamProperty {   name: string;   operator: \&quot;contains\&quot; | \&quot;eq\&quot; | \&quot;bool\&quot; | \&quot;gt\&quot; | \&quot;lt\&quot;;   value: string; }  interface ParamOperator {   name: \&quot;and\&quot; | \&quot;or\&quot;; }  interface Filter {   type: \&quot;property\&quot; | \&quot;operator\&quot;;   param: ParamOperator | ParamProperty; }  interface FilterList {   Filters: Array&lt;Filter&gt;; }  &#x60;&#x60;&#x60;  ## Examples  This is a example to filter and get only the resource what property \&quot;confirmed\&quot; is \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {   \&quot;type\&quot;: \&quot;property\&quot;,   \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;confirmed\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;       }   } ] &#x60;&#x60;&#x60;  This one, filter resource by the property \&quot;id\&quot; inside \&quot;info\&quot; structure when it is equal to \&quot;manjaro\&quot; and online property is set to \&quot;true\&quot; &#x60;&#x60;&#x60;json [   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;info.id\&quot;,       \&quot;operator\&quot;: \&quot;eq\&quot;,       \&quot;value\&quot;: \&quot;manjaro\&quot;     }   },   {     \&quot;type\&quot;: \&quot;property\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;online\&quot;,       \&quot;operator\&quot;: \&quot;bool\&quot;,       \&quot;value\&quot;: \&quot;true\&quot;     }   },   {     \&quot;type\&quot;: \&quot;operator\&quot;,     \&quot;params\&quot;: {       \&quot;name\&quot;: \&quot;and\&quot;     }   } ] &#x60;&#x60;&#x60;  | (optional) defaults to undefined|
| **page** | [**number**] | Page number | (optional) defaults to 1|
| **perPage** | [**number**] | Items per page | (optional) defaults to 10|


### Return type

**Array<UserAdminResponse>**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get a list of users. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAnnouncementsAdmin**
> Array<AnnouncementShort> listAnnouncementsAdmin()

List the announcements posted by ShellHub Cloud.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)
let orderBy: ListAnnouncementsOrderByParameter; // (optional) (default to undefined)

const { status, data } = await apiInstance.listAnnouncementsAdmin(
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

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to get the announcements. |  * X-Total-Count - Announcements\&#39; total number. <br>  |
|**400** | Bad request |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **loginAdmin**
> LoginAdmin200Response loginAdmin()

Login on Admin

### Example

```typescript
import {
    AdminApi,
    Configuration,
    LoginAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let loginAdminRequest: LoginAdminRequest; // (optional)

const { status, data } = await apiInstance.loginAdmin(
    loginAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginAdminRequest** | **LoginAdminRequest**|  | |


### Return type

**LoginAdmin200Response**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to Login on Admin |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **sendLicense**
> sendLicense()

Send license data

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

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

# **setSessionAuthenticationStatusAdmin**
> setSessionAuthenticationStatusAdmin()

Set session authentication status.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    SetSessionAuthenticationStatusRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; // (default to undefined)
let setSessionAuthenticationStatusRequest: SetSessionAuthenticationStatusRequest; // (optional)

const { status, data } = await apiInstance.setSessionAuthenticationStatusAdmin(
    uid,
    setSessionAuthenticationStatusRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **setSessionAuthenticationStatusRequest** | **SetSessionAuthenticationStatusRequest**|  | |
| **uid** | [**string**] |  | defaults to undefined|


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
|**200** | Success |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateAnnouncement**
> Announcement updateAnnouncement()

Update an announcement.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    CreateAnnouncementRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uuid: string; // (default to undefined)
let createAnnouncementRequest: CreateAnnouncementRequest; // (optional)

const { status, data } = await apiInstance.updateAnnouncement(
    uuid,
    createAnnouncementRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **createAnnouncementRequest** | **CreateAnnouncementRequest**|  | |
| **uuid** | [**string**] |  | defaults to undefined|


### Return type

**Announcement**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update an announcement. |  -  |
|**401** | Unauthorized |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceNameAdmin**
> updateDeviceNameAdmin()

Update device\'s name.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    UpdateDeviceNameAdminRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let updateDeviceNameAdminRequest: UpdateDeviceNameAdminRequest; // (optional)

const { status, data } = await apiInstance.updateDeviceNameAdmin(
    uid,
    updateDeviceNameAdminRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **updateDeviceNameAdminRequest** | **UpdateDeviceNameAdminRequest**|  | |
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
|**200** | Success to update device\&#39;s name. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**409** | Conflict |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateDeviceStatusAdmin**
> updateDeviceStatusAdmin()

Update device\'s status.

### Example

```typescript
import {
    AdminApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let uid: string; //Device\'s UID (default to undefined)
let status: DeviceStatus; //Device\'s status (default to undefined)

const { status, data } = await apiInstance.updateDeviceStatusAdmin(
    uid,
    status
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **uid** | [**string**] | Device\&#39;s UID | defaults to undefined|
| **status** | **DeviceStatus** | Device\&#39;s status | defaults to undefined|


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
|**200** | Success to update device status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**402** | Payment required. |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateFirewallRuleAdmin**
> FirewallRulesResponse updateFirewallRuleAdmin()

Update a firewall rule.

### Example

```typescript
import {
    AdminApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AdminApi(configuration);

let id: string; // (default to undefined)
let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.updateFirewallRuleAdmin(
    id,
    firewallRulesRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **firewallRulesRequest** | **FirewallRulesRequest**|  | |
| **id** | [**string**] |  | defaults to undefined|


### Return type

**FirewallRulesResponse**

### Authorization

[api-key](../README.md#api-key), [jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to update firewall rule. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

