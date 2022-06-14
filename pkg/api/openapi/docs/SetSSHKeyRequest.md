# SetSSHKeyRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Fingerprint** | Pointer to **string** | Device fingerprint | [optional] 
**Data** | Pointer to **string** | Device SSH public key | [optional] 

## Methods

### NewSetSSHKeyRequest

`func NewSetSSHKeyRequest() *SetSSHKeyRequest`

NewSetSSHKeyRequest instantiates a new SetSSHKeyRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSetSSHKeyRequestWithDefaults

`func NewSetSSHKeyRequestWithDefaults() *SetSSHKeyRequest`

NewSetSSHKeyRequestWithDefaults instantiates a new SetSSHKeyRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFingerprint

`func (o *SetSSHKeyRequest) GetFingerprint() string`

GetFingerprint returns the Fingerprint field if non-nil, zero value otherwise.

### GetFingerprintOk

`func (o *SetSSHKeyRequest) GetFingerprintOk() (*string, bool)`

GetFingerprintOk returns a tuple with the Fingerprint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFingerprint

`func (o *SetSSHKeyRequest) SetFingerprint(v string)`

SetFingerprint sets Fingerprint field to given value.

### HasFingerprint

`func (o *SetSSHKeyRequest) HasFingerprint() bool`

HasFingerprint returns a boolean if a field has been set.

### GetData

`func (o *SetSSHKeyRequest) GetData() string`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *SetSSHKeyRequest) GetDataOk() (*string, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *SetSSHKeyRequest) SetData(v string)`

SetData sets Data field to given value.

### HasData

`func (o *SetSSHKeyRequest) HasData() bool`

HasData returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


