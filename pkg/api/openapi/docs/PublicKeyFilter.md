# PublicKeyFilter

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Hostname** | **string** | Public key&#39;s regex hostname. | 
**Tags** | **[]string** | Public key&#39;s tags. | 

## Methods

### NewPublicKeyFilter

`func NewPublicKeyFilter(hostname string, tags []string, ) *PublicKeyFilter`

NewPublicKeyFilter instantiates a new PublicKeyFilter object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPublicKeyFilterWithDefaults

`func NewPublicKeyFilterWithDefaults() *PublicKeyFilter`

NewPublicKeyFilterWithDefaults instantiates a new PublicKeyFilter object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHostname

`func (o *PublicKeyFilter) GetHostname() string`

GetHostname returns the Hostname field if non-nil, zero value otherwise.

### GetHostnameOk

`func (o *PublicKeyFilter) GetHostnameOk() (*string, bool)`

GetHostnameOk returns a tuple with the Hostname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostname

`func (o *PublicKeyFilter) SetHostname(v string)`

SetHostname sets Hostname field to given value.


### GetTags

`func (o *PublicKeyFilter) GetTags() []string`

GetTags returns the Tags field if non-nil, zero value otherwise.

### GetTagsOk

`func (o *PublicKeyFilter) GetTagsOk() (*[]string, bool)`

GetTagsOk returns a tuple with the Tags field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTags

`func (o *PublicKeyFilter) SetTags(v []string)`

SetTags sets Tags field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


