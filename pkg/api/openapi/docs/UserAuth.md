# UserAuth

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Token** | Pointer to **string** | JWT Token | [optional] 
**User** | Pointer to **string** | User&#39;s name. | [optional] 
**Name** | Pointer to **string** | User&#39;s display name. | [optional] 
**Id** | Pointer to **string** | User&#39;s ID. | [optional] 
**Tenant** | Pointer to **string** | Namespace&#39;s tenant ID | [optional] 
**Role** | Pointer to [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] 
**Email** | Pointer to **string** | User&#39;s E-mail. | [optional] 

## Methods

### NewUserAuth

`func NewUserAuth() *UserAuth`

NewUserAuth instantiates a new UserAuth object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserAuthWithDefaults

`func NewUserAuthWithDefaults() *UserAuth`

NewUserAuthWithDefaults instantiates a new UserAuth object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToken

`func (o *UserAuth) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *UserAuth) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *UserAuth) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *UserAuth) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetUser

`func (o *UserAuth) GetUser() string`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *UserAuth) GetUserOk() (*string, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *UserAuth) SetUser(v string)`

SetUser sets User field to given value.

### HasUser

`func (o *UserAuth) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetName

`func (o *UserAuth) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UserAuth) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UserAuth) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UserAuth) HasName() bool`

HasName returns a boolean if a field has been set.

### GetId

`func (o *UserAuth) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UserAuth) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UserAuth) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *UserAuth) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTenant

`func (o *UserAuth) GetTenant() string`

GetTenant returns the Tenant field if non-nil, zero value otherwise.

### GetTenantOk

`func (o *UserAuth) GetTenantOk() (*string, bool)`

GetTenantOk returns a tuple with the Tenant field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenant

`func (o *UserAuth) SetTenant(v string)`

SetTenant sets Tenant field to given value.

### HasTenant

`func (o *UserAuth) HasTenant() bool`

HasTenant returns a boolean if a field has been set.

### GetRole

`func (o *UserAuth) GetRole() NamespaceMemberRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *UserAuth) GetRoleOk() (*NamespaceMemberRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *UserAuth) SetRole(v NamespaceMemberRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *UserAuth) HasRole() bool`

HasRole returns a boolean if a field has been set.

### GetEmail

`func (o *UserAuth) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserAuth) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserAuth) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *UserAuth) HasEmail() bool`

HasEmail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


