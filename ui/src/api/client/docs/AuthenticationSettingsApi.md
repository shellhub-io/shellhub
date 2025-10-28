# AuthenticationSettingsApi

All URIs are relative to *http://localhost*

|Method | HTTP request | Description|
|------------- | ------------- | -------------|
|[**configureLocalAuthentication**](#configurelocalauthentication) | **PUT** /admin/api/authentication/local | Configure Local Authentication|
|[**configureSAMLAuthentication**](#configuresamlauthentication) | **PUT** /admin/api/authentication/saml | Configure SAML Authentication|
|[**getAuthenticationSettings**](#getauthenticationsettings) | **GET** /admin/api/authentication | Get Authentication Settings|

# **configureLocalAuthentication**
> configureLocalAuthentication(configureLocalAuthenticationRequest)

Configure local authentication settings for the ShellHub instance.

### Example

```typescript
import {
    AuthenticationSettingsApi,
    Configuration,
    ConfigureLocalAuthenticationRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthenticationSettingsApi(configuration);

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
    AuthenticationSettingsApi,
    Configuration,
    ConfigureSAMLAuthenticationRequest
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthenticationSettingsApi(configuration);

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

# **getAuthenticationSettings**
> GetAuthenticationSettings200Response getAuthenticationSettings()

Retrieves the current authentication settings.

### Example

```typescript
import {
    AuthenticationSettingsApi,
    Configuration
} from './api';

const configuration = new Configuration();
const apiInstance = new AuthenticationSettingsApi(configuration);

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

