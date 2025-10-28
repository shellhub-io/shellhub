# Info


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**version** | **string** | The current version of ShellHub. | [optional] [default to undefined]
**endpoints** | [**InfoEndpoints**](InfoEndpoints.md) |  | [optional] [default to undefined]
**setup** | **boolean** | Indicates whether the instance setup is complete. | [optional] [default to undefined]
**authentication** | [**InfoAuthentication**](InfoAuthentication.md) |  | [optional] [default to undefined]

## Example

```typescript
import { Info } from './api';

const instance: Info = {
    version,
    endpoints,
    setup,
    authentication,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
