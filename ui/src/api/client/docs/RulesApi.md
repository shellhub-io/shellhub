# RulesApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**createFirewallRule**](#createfirewallrule) | **POST** /api/firewall/rules | Create firewall rule|
|[**createFirewallRuleAdmin**](#createfirewallruleadmin) | **POST** /admin/api/firewall/rules | Create firewall rule Admin|
|[**deleteFirewallRule**](#deletefirewallrule) | **DELETE** /api/firewall/rules/{id} | Delete firewall rule|
|[**deleteFirewallRuleAdmin**](#deletefirewallruleadmin) | **DELETE** /admin/api/firewall/rules/{id} | Delete firewall rule admin|
|[**getFirewallRule**](#getfirewallrule) | **GET** /api/firewall/rules/{id} | Get firewall rule|
|[**getFirewallRuleAdmin**](#getfirewallruleadmin) | **GET** /admin/api/firewall/rules/{id} | Get firewall rule admin|
|[**getFirewallRules**](#getfirewallrules) | **GET** /api/firewall/rules | Get firewall rules|
|[**getFirewallRulesAdmin**](#getfirewallrulesadmin) | **GET** /admin/api/firewall/rules | Get firewall rules Admin|
|[**updateFirewallRule**](#updatefirewallrule) | **PUT** /api/firewall/rules/{id} | Update firewall rule|
|[**updateFirewallRuleAdmin**](#updatefirewallruleadmin) | **PUT** /admin/api/firewall/rules/{id} | Update firewall rule admin|

# **createFirewallRule**
> FirewallRulesResponse createFirewallRule()

Create a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.createFirewallRule(
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

# **createFirewallRuleAdmin**
> FirewallRulesResponse createFirewallRuleAdmin()

Create a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

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

# **deleteFirewallRule**
> deleteFirewallRule()

Delete a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.deleteFirewallRule(
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

# **deleteFirewallRuleAdmin**
> deleteFirewallRuleAdmin()

Delete a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

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

# **getFirewallRule**
> FirewallRulesResponse getFirewallRule()

Get a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

let id: string; // (default to undefined)

const { status, data } = await apiInstance.getFirewallRule(
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

# **getFirewallRuleAdmin**
> FirewallRulesResponse getFirewallRuleAdmin()

Get a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

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

# **getFirewallRules**
> Array<FirewallRulesResponse> getFirewallRules()

Get a list of firewall rules.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

let page: number; //Page number (optional) (default to 1)
let perPage: number; //Items per page (optional) (default to 10)

const { status, data } = await apiInstance.getFirewallRules(
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
|**200** | Success to get firewall rules. |  * X-Total-Count -  <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getFirewallRulesAdmin**
> Array<FirewallRulesResponse> getFirewallRulesAdmin()

Get a list of firewall rules.

### Example

```typescript
import {
    RulesApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

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
|**200** | Success to get firewall rules. |  * X-Total-Count -  <br>  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateFirewallRule**
> FirewallRulesResponse updateFirewallRule()

Update a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

let id: string; // (default to undefined)
let firewallRulesRequest: FirewallRulesRequest; // (optional)

const { status, data } = await apiInstance.updateFirewallRule(
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

# **updateFirewallRuleAdmin**
> FirewallRulesResponse updateFirewallRuleAdmin()

Update a firewall rule.

### Example

```typescript
import {
    RulesApi,
    Configuration,
    FirewallRulesRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new RulesApi(configuration);

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

