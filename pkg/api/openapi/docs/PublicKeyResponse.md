# PublicKeyResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Data** | Pointer to **string** | Public key&#39;s data.  The &#x60;data&#x60; field receives the public key enconded as &#x60;base64&#x60; string.  | [optional] 
**Fingerprint** | Pointer to **string** | Public key&#39;s fingerprint. | [optional] 
**CreatedAt** | Pointer to **time.Time** | Public key&#39;s creation date. | [optional] 
**TenantId** | Pointer to **string** | Namespace&#39;s tenant ID | [optional] 
**Name** | Pointer to **string** | Public key&#39;s name. | [optional] 
**Filter** | Pointer to [**PublicKeyFilter**](PublicKeyFilter.md) |  | [optional] 
**Username** | Pointer to **string** | Public key&#39;s regex username.   The &#x60;username&#x60; field define which user, in the device, may be access through this public key.  | [optional] 

## Methods

### NewPublicKeyResponse

`func NewPublicKeyResponse() *PublicKeyResponse`

NewPublicKeyResponse instantiates a new PublicKeyResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPublicKeyResponseWithDefaults

`func NewPublicKeyResponseWithDefaults() *PublicKeyResponse`

NewPublicKeyResponseWithDefaults instantiates a new PublicKeyResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetData

`func (o *PublicKeyResponse) GetData() string`

GetData returns the Data field if non-nil, zero value otherwise.

### GetDataOk

`func (o *PublicKeyResponse) GetDataOk() (*string, bool)`

GetDataOk returns a tuple with the Data field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetData

`func (o *PublicKeyResponse) SetData(v string)`

SetData sets Data field to given value.

### HasData

`func (o *PublicKeyResponse) HasData() bool`

HasData returns a boolean if a field has been set.

### GetFingerprint

`func (o *PublicKeyResponse) GetFingerprint() string`

GetFingerprint returns the Fingerprint field if non-nil, zero value otherwise.

### GetFingerprintOk

`func (o *PublicKeyResponse) GetFingerprintOk() (*string, bool)`

GetFingerprintOk returns a tuple with the Fingerprint field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFingerprint

`func (o *PublicKeyResponse) SetFingerprint(v string)`

SetFingerprint sets Fingerprint field to given value.

### HasFingerprint

`func (o *PublicKeyResponse) HasFingerprint() bool`

HasFingerprint returns a boolean if a field has been set.

### GetCreatedAt

`func (o *PublicKeyResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PublicKeyResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PublicKeyResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *PublicKeyResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetTenantId

`func (o *PublicKeyResponse) GetTenantId() string`

GetTenantId returns the TenantId field if non-nil, zero value otherwise.

### GetTenantIdOk

`func (o *PublicKeyResponse) GetTenantIdOk() (*string, bool)`

GetTenantIdOk returns a tuple with the TenantId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTenantId

`func (o *PublicKeyResponse) SetTenantId(v string)`

SetTenantId sets TenantId field to given value.

### HasTenantId

`func (o *PublicKeyResponse) HasTenantId() bool`

HasTenantId returns a boolean if a field has been set.

### GetName

`func (o *PublicKeyResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PublicKeyResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PublicKeyResponse) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PublicKeyResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetFilter

`func (o *PublicKeyResponse) GetFilter() PublicKeyFilter`

GetFilter returns the Filter field if non-nil, zero value otherwise.

### GetFilterOk

`func (o *PublicKeyResponse) GetFilterOk() (*PublicKeyFilter, bool)`

GetFilterOk returns a tuple with the Filter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFilter

`func (o *PublicKeyResponse) SetFilter(v PublicKeyFilter)`

SetFilter sets Filter field to given value.

### HasFilter

`func (o *PublicKeyResponse) HasFilter() bool`

HasFilter returns a boolean if a field has been set.

### GetUsername

`func (o *PublicKeyResponse) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *PublicKeyResponse) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *PublicKeyResponse) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *PublicKeyResponse) HasUsername() bool`

HasUsername returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


