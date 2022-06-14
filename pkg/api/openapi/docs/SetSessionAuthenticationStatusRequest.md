# SetSessionAuthenticationStatusRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Authenticated** | Pointer to **bool** | Session&#39;s authentication status. | [optional] 

## Methods

### NewSetSessionAuthenticationStatusRequest

`func NewSetSessionAuthenticationStatusRequest() *SetSessionAuthenticationStatusRequest`

NewSetSessionAuthenticationStatusRequest instantiates a new SetSessionAuthenticationStatusRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetSessionAuthenticationStatusRequestWithDefaults

`func NewSetSessionAuthenticationStatusRequestWithDefaults() *SetSessionAuthenticationStatusRequest`

NewSetSessionAuthenticationStatusRequestWithDefaults instantiates a new SetSessionAuthenticationStatusRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAuthenticated

`func (o *SetSessionAuthenticationStatusRequest) GetAuthenticated() bool`

GetAuthenticated returns the Authenticated field if non-nil, zero value otherwise.

### GetAuthenticatedOk

`func (o *SetSessionAuthenticationStatusRequest) GetAuthenticatedOk() (*bool, bool)`

GetAuthenticatedOk returns a tuple with the Authenticated field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthenticated

`func (o *SetSessionAuthenticationStatusRequest) SetAuthenticated(v bool)`

SetAuthenticated sets Authenticated field to given value.

### HasAuthenticated

`func (o *SetSessionAuthenticationStatusRequest) HasAuthenticated() bool`

HasAuthenticated returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


