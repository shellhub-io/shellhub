# InfoEndpoints

Network endpoints for the ShellHub instance.

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ssh** | **string** | The SSH endpoint where devices connect. | [optional] [default to undefined]
**api** | **string** | The API endpoint for managing ShellHub configurations. | [optional] [default to undefined]

## Example

```typescript
import { InfoEndpoints } from './api';

const instance: InfoEndpoints = {
    ssh,
    api,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
