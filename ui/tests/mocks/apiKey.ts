import { IApiKey } from "@/interfaces/IApiKey";

export const mockApiKey: IApiKey = {
  id: "generated-api-key-123",
  tenant_id: "00000-0000-0000-0000-00000000000",
  name: "test-key",
  role: "administrator",
  expires_in: 30,
};

export const mockApiKeys: IApiKey[] = [
  mockApiKey,
  {
    ...mockApiKey,
    id: "generated-api-key-456",
    name: "another-key",
    role: "observer",
    expires_in: -1,
  },
];
