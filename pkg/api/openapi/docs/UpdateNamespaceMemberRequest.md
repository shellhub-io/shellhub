# UpdateNamespaceMemberRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Role** | Pointer to [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] 

## Methods

### NewUpdateNamespaceMemberRequest

`func NewUpdateNamespaceMemberRequest() *UpdateNamespaceMemberRequest`

NewUpdateNamespaceMemberRequest instantiates a new UpdateNamespaceMemberRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdateNamespaceMemberRequestWithDefaults

`func NewUpdateNamespaceMemberRequestWithDefaults() *UpdateNamespaceMemberRequest`

NewUpdateNamespaceMemberRequestWithDefaults instantiates a new UpdateNamespaceMemberRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRole

`func (o *UpdateNamespaceMemberRequest) GetRole() NamespaceMemberRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *UpdateNamespaceMemberRequest) GetRoleOk() (*NamespaceMemberRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *UpdateNamespaceMemberRequest) SetRole(v NamespaceMemberRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *UpdateNamespaceMemberRequest) HasRole() bool`

HasRole returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


