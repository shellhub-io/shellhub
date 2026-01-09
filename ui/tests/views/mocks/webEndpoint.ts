import { IWebEndpoint } from "@/interfaces/IWebEndpoints";
import { mockDevice } from "@tests/views/mocks/device";

/**
 * Mock web endpoint data for testing.
 * Provides a complete web endpoint object with all required fields.
 */
export const mockWebEndpoint: IWebEndpoint = {
  address: "123abc",
  full_address: "localhost:8080",
  device_uid: "a582b47a42d",
  device: mockDevice,
  host: "localhost",
  port: 8080,
  expires_in: "2099-01-01T00:00:00Z",
};

/**
 * Mock web endpoints array for testing lists.
 * Provides multiple web endpoints for list/table testing scenarios.
 */
export const mockWebEndpoints: IWebEndpoint[] = [mockWebEndpoint];
