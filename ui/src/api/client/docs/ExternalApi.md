# ExternalApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**authUser**](#authuser) | **POST** /api/auth/user | Auth a user|
|[**login**](#login) | **POST** /api/login | Login|

# **authUser**
> UserAuth authUser()

Authenticate a user, returning the session\'s JWT token and data about the user.

### Example

```typescript
import {
    ExternalApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new ExternalApi(configuration);

let loginRequest: LoginRequest; // (optional)

const { status, data } = await apiInstance.authUser(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


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
|**200** | Success to auth the user. |  -  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  -  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **login**
> UserAuth login()

Authenticate a \"local\" user by returning the session\'s JWT token and user data. Local users are those registered via the ShellHub form without relying on external Identity Providers (IdPs).  Authentication may result in an account lockout after N consecutive incorrect login attempts. The lockout applies specifically to a particular source and user combination. Check for the presence of the `X-Account-Lockout` header to determine the account lockout status. When it\'s 0, there are no active lockouts.  Users with MFA enabled cannot authenticate via this route. In such cases, the API will respond with a status `401` and an `X-MFA-Token` header with a UUID. Authentication must be med to `/api/mfa/auth` with this token in these instances. 

### Example

```typescript
import {
    ExternalApi,
    Configuration,
    LoginRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new ExternalApi(configuration);

let loginRequest: LoginRequest; // (optional)

const { status, data } = await apiInstance.login(
    loginRequest
);
```

### Parameters

|Name | Type | Description  | Notes|
|------------- | ------------- | ------------- | -------------|
| **loginRequest** | **LoginRequest**|  | |


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
|**200** | Success to login. |  * X-Account-Lockout -  <br>  * X-MFA-Token -  <br>  |
|**400** | Bad request |  -  |
|**401** | Unauthorized |  * X-MFA-Token -  <br>  |
|**403** | Forbidden |  -  |
|**404** | Not found |  -  |
|**429** | Too Many Requests |  * X-Account-Lockout -  <br>  * X-MFA-Token -  <br>  |
|**500** | Internal error |  -  |

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

