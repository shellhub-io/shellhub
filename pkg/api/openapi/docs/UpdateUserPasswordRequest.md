# UpdateUserPasswordRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CurrentPassword** | Pointer to **string** | User current password | [optional] 
**NewPassword** | Pointer to **string** | User new password | [optional] 

## Methods

### NewUpdateUserPasswordRequest

`func NewUpdateUserPasswordRequest() *UpdateUserPasswordRequest`

NewUpdateUserPasswordRequest instantiates a new UpdateUserPasswordRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateUserPasswordRequestWithDefaults

`func NewUpdateUserPasswordRequestWithDefaults() *UpdateUserPasswordRequest`

NewUpdateUserPasswordRequestWithDefaults instantiates a new UpdateUserPasswordRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCurrentPassword

`func (o *UpdateUserPasswordRequest) GetCurrentPassword() string`

GetCurrentPassword returns the CurrentPassword field if non-nil, zero value otherwise.

### GetCurrentPasswordOk

`func (o *UpdateUserPasswordRequest) GetCurrentPasswordOk() (*string, bool)`

GetCurrentPasswordOk returns a tuple with the CurrentPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentPassword

`func (o *UpdateUserPasswordRequest) SetCurrentPassword(v string)`

SetCurrentPassword sets CurrentPassword field to given value.

### HasCurrentPassword

`func (o *UpdateUserPasswordRequest) HasCurrentPassword() bool`

HasCurrentPassword returns a boolean if a field has been set.

### GetNewPassword

`func (o *UpdateUserPasswordRequest) GetNewPassword() string`

GetNewPassword returns the NewPassword field if non-nil, zero value otherwise.

### GetNewPasswordOk

`func (o *UpdateUserPasswordRequest) GetNewPasswordOk() (*string, bool)`

GetNewPasswordOk returns a tuple with the NewPassword field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNewPassword

`func (o *UpdateUserPasswordRequest) SetNewPassword(v string)`

SetNewPassword sets NewPassword field to given value.

### HasNewPassword

`func (o *UpdateUserPasswordRequest) HasNewPassword() bool`

HasNewPassword returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


