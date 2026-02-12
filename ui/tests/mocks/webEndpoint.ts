import { IWebEndpoint } from "@/interfaces/IWebEndpoints";
import { mockDevice } from "./device";

/**
 * Mock web endpoint data for testing.
 * Provides a complete web endpoint object with all required fields.
 */
export const mockWebEndpoint: IWebEndpoint = {
  address: "endpoint-123",
  full_address: "endpoint-123.example.com",
  device_uid: "device-123",
  device: mockDevice,
  host: "192.168.1.1",
  port: 8080,
  expires_in: "2099-12-31T23:59:59Z",
};

/**
 * Mock expired web endpoint for testing expiration scenarios.
 */
export const mockExpiredWebEndpoint: IWebEndpoint = {
  ...mockWebEndpoint,
  address: "endpoint-expired",
  expires_in: "2020-01-01T00:00:00Z",
};

/**
 * Mock web endpoint that never expires.
 */
export const mockNeverExpiresWebEndpoint: IWebEndpoint = {
  ...mockWebEndpoint,
  address: "endpoint-never",
  expires_in: "0001-01-01T00:00:00Z",
};

/**
 * Mock web endpoint with TLS enabled.
 */
export const mockWebEndpointWithTLS: IWebEndpoint = {
  ...mockWebEndpoint,
  address: "endpoint-tls",
  tls: {
    enabled: true,
    domain: "secure.example.com",
    verify: true,
  },
};

/**
 * Mock web endpoint with TLS disabled.
 */
export const mockWebEndpointWithoutTLS: IWebEndpoint = {
  ...mockWebEndpoint,
  address: "endpoint-no-tls",
  tls: {
    enabled: false,
    domain: "",
    verify: false,
  },
};

/**
 * Mock web endpoints array for testing lists.
 * Provides multiple web endpoints for list/table testing scenarios.
 */
export const mockWebEndpoints: IWebEndpoint[] = [
  mockWebEndpoint,
  mockExpiredWebEndpoint,
  mockWebEndpointWithTLS,
];
