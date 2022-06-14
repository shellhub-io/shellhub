# NamespaceMembersInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | User&#39;s ID. | [optional] 
**Role** | Pointer to [**NamespaceMemberRole**](NamespaceMemberRole.md) |  | [optional] 

## Methods

### NewNamespaceMembersInner

`func NewNamespaceMembersInner() *NamespaceMembersInner`

NewNamespaceMembersInner instantiates a new NamespaceMembersInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNamespaceMembersInnerWithDefaults

`func NewNamespaceMembersInnerWithDefaults() *NamespaceMembersInner`

NewNamespaceMembersInnerWithDefaults instantiates a new NamespaceMembersInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *NamespaceMembersInner) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *NamespaceMembersInner) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *NamespaceMembersInner) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *NamespaceMembersInner) HasId() bool`

HasId returns a boolean if a field has been set.

### GetRole

`func (o *NamespaceMembersInner) GetRole() NamespaceMemberRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *NamespaceMembersInner) GetRoleOk() (*NamespaceMemberRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *NamespaceMembersInner) SetRole(v NamespaceMemberRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *NamespaceMembersInner) HasRole() bool`

HasRole returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


