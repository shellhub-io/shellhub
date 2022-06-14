# AddNamespaceMemberRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Username** | **string** | Member&#39;s username. | 
**Role** | [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | 

## Methods

### NewAddNamespaceMemberRequest

`func NewAddNamespaceMemberRequest(username string, role NamespaceMemberRole, ) *AddNamespaceMemberRequest`

NewAddNamespaceMemberRequest instantiates a new AddNamespaceMemberRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddNamespaceMemberRequestWithDefaults

`func NewAddNamespaceMemberRequestWithDefaults() *AddNamespaceMemberRequest`

NewAddNamespaceMemberRequestWithDefaults instantiates a new AddNamespaceMemberRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *AddNamespaceMemberRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *AddNamespaceMemberRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *AddNamespaceMemberRequest) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetRole

`func (o *AddNamespaceMemberRequest) GetRole() NamespaceMemberRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *AddNamespaceMemberRequest) GetRoleOk() (*NamespaceMemberRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *AddNamespaceMemberRequest) SetRole(v NamespaceMemberRole)`

SetRole sets Role field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


