# NamespaceSettings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SessionRecord** | Pointer to **bool** | Namespace&#39;s session record status | [optional] [default to true]

## Methods

### NewNamespaceSettings

`func NewNamespaceSettings() *NamespaceSettings`

NewNamespaceSettings instantiates a new NamespaceSettings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNamespaceSettingsWithDefaults

`func NewNamespaceSettingsWithDefaults() *NamespaceSettings`

NewNamespaceSettingsWithDefaults instantiates a new NamespaceSettings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSessionRecord

`func (o *NamespaceSettings) GetSessionRecord() bool`

GetSessionRecord returns the SessionRecord field if non-nil, zero value otherwise.

### GetSessionRecordOk

`func (o *NamespaceSettings) GetSessionRecordOk() (*bool, bool)`

GetSessionRecordOk returns a tuple with the SessionRecord field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSessionRecord

`func (o *NamespaceSettings) SetSessionRecord(v bool)`

SetSessionRecord sets SessionRecord field to given value.

### HasSessionRecord

`func (o *NamespaceSettings) HasSessionRecord() bool`

HasSessionRecord returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


