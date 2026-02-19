# Tag

A tag represents a label or category that can be attached to devices, firewall rules and public keys for organization and filtering purposes. 

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**name** | **string** | The display name of the tag | [default to undefined]
**tenant_id** | **string** | The tenant ID that owns this tag | [default to undefined]
**created_at** | **string** | The timestamp when the tag was created | [default to undefined]
**updated_at** | **string** | The timestamp when the tag was last updated | [default to undefined]

## Example

```typescript
import { Tag } from './api';

const instance: Tag = {
    name,
    tenant_id,
    created_at,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
