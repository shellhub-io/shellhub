# UpdatePublicKeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Public key&#39;s name. | 
**Username** | **string** | Public key&#39;s username. | 
**Filter** | [**PublicKeyFilter**](PublicKeyFilter.md) |  | 

## Methods

### NewUpdatePublicKeyRequest

`func NewUpdatePublicKeyRequest(name string, username string, filter PublicKeyFilter, ) *UpdatePublicKeyRequest`

NewUpdatePublicKeyRequest instantiates a new UpdatePublicKeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUpdatePublicKeyRequestWithDefaults

`func NewUpdatePublicKeyRequestWithDefaults() *UpdatePublicKeyRequest`

NewUpdatePublicKeyRequestWithDefaults instantiates a new UpdatePublicKeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *UpdatePublicKeyRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UpdatePublicKeyRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UpdatePublicKeyRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUsername

`func (o *UpdatePublicKeyRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *UpdatePublicKeyRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *UpdatePublicKeyRequest) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetFilter

`func (o *UpdatePublicKeyRequest) GetFilter() PublicKeyFilter`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *UpdatePublicKeyRequest) GetFilterOk() (*PublicKeyFilter, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *UpdatePublicKeyRequest) SetFilter(v PublicKeyFilter)`

SetFilter sets Filter field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


