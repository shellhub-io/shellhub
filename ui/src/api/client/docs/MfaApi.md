# MfaApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**authMFA**](#authmfa) | **POST** /api/user/mfa/auth | Auth MFA|
|[**disableMFA**](#disablemfa) | **PUT** /api/user/mfa/disable | Disable MFA|
|[**enableMFA**](#enablemfa) | **PUT** /api/user/mfa/enable | Enable MFA|
|[**generateMFA**](#generatemfa) | **GET** /api/user/mfa/generate | Generate MFA Credentials|
|[**mfaRecover**](#mfarecover) | **POST** /api/user/mfa/recover | Recover MFA|
|[**requestResetMFA**](#requestresetmfa) | **POST** /api/user/mfa/reset | Request Reset MFA|
|[**resetMFA**](#resetmfa) | **PUT** /api/user/mfa/reset/{user-id} | Reset MFA|

# **authMFA**
> UserAuth authMFA()

Authenticate a user who has MFA enabled. This endpoint should be called after the default authUser endpoint, which generates an `X-MFA-Token` indicating that the user has already authenticated with a password. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    MfaAuth
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let mfaAuth: MfaAuth; // (optional)

const { status, data } = await apiInstance.authMFA(
    mfaAuth
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaAuth** | **MfaAuth**|  | |


### Return type

**UserAuth**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to authenticate. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **disableMFA**
> disableMFA()

Disable MFA for a user. To disable MFA, the user must provide either a recovery code or the current MFA code. If a recovery code is used, it will be invalidated for future use.  The recovery code used to regain access to the account can be used within a 10-minute window on this endpoint. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    MfaDisable
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let mfaDisable: MfaDisable; // (optional)

const { status, data } = await apiInstance.disableMFA(
    mfaDisable
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaDisable** | **MfaDisable**|  | |


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
|**200** | Success to disable a MFA status. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **enableMFA**
> enableMFA()

Enable MFA for a user. The secret and recovery codes must be created by the generateMFA endpoint. Users with MFA already enabled cannot override their MFA credentials; in these cases, a user must disable MFA before proceeding. The recovery e-mail must be a valid value in order to enable the MFA. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    MfaEnable
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let mfaEnable: MfaEnable; // (optional)

const { status, data } = await apiInstance.enableMFA(
    mfaEnable
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaEnable** | **MfaEnable**|  | |


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
|**200** | Success to enable a MFA. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **generateMFA**
> MfaGenerate generateMFA()

Generate the credentials to enable a user\'s MFA. The user must save the recovery codes a secure manner. 

### Example

```typescript
import {
    MfaApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

const { status, data } = await apiInstance.generateMFA();
```

### Parameters
This endpoint does not have any parameters.


### Return type

**MfaGenerate**

### Authorization

[jwt](../README.md#jwt)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Successfully generated MFA credentials. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **mfaRecover**
> UserAuth mfaRecover()

Recover account access by providing one of the user\'s recovery codes. It will be invalidated for future uses.  The recovery code will be cached for 10 minutes. During this period, the user can use the same recovery code to disable their MFA without needing to provide two separate codes. The `X-Expires-At` header specifies the epoch value marking the end of the cache period. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    MfaRecover
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let mfaRecover: MfaRecover; // (optional)

const { status, data } = await apiInstance.mfaRecover(
    mfaRecover
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaRecover** | **MfaRecover**|  | |


### Return type

**UserAuth**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success recover account\&#39;s access. |  * X-Expires-At - The epoch time at which the recovery code will become invalid. <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **requestResetMFA**
> requestResetMFA()

Sends an email to both the user\'s main and recovery addresses. Each email contains a unique code, which remains valid for at most 1 day. The user must provide both codes to reset their MFA. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    RequestResetMFARequest
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let requestResetMFARequest: RequestResetMFARequest; // (optional)

const { status, data } = await apiInstance.requestResetMFA(
    requestResetMFARequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **requestResetMFARequest** | **RequestResetMFARequest**|  | |


### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to send the email. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **resetMFA**
> UserAuth resetMFA()

Similar to the `disableMFA` operation, this endpoint uses the two codes sent by `requestResetMFA` instead of a TOTP or recovery code. The user ID must be the same as the one used for `requestResetMFA`. 

### Example

```typescript
import {
    MfaApi,
    Configuration,
    MfaReset
} from './api';

const configuration = new Configuration();
const apiInstance = new MfaApi(configuration);

let userId: string; // (default to undefined)
let mfaReset: MfaReset; // (optional)

const { status, data } = await apiInstance.resetMFA(
    userId,
    mfaReset
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **mfaReset** | **MfaReset**|  | |
| **userId** | [**string**] |  | defaults to undefined|


### Return type

**UserAuth**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json


### HTTP response details
| Status code | Description | Response headers |
|-------------|-------------|------------------|
|**200** | Success to reset and authenticate. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

