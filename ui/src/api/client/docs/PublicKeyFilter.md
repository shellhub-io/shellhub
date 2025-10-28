# PublicKeyFilter

Public key\'s filter rule.   The `filter`` rule defines how if the public key is valid to a device.  - When `hostname` object is set, the public key will be used in a device what matches with hostname. - When `tags` object is set, it matches the device what contains at least one of that tags. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**hostname** | **string** | Public key\&#39;s regex hostname. | [default to undefined]
**tags** | [**Set&lt;Tag&gt;**](Tag.md) | Public key\&#39;s tags. | [default to undefined]

## Example

```typescript
import { PublicKeyFilter } from './api';

const instance: PublicKeyFilter = {
    hostname,
    tags,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
