/**
 * Centralized mock data exports for view tests.
 *
 * This module provides all common mock objects used across view tests,
 * ensuring consistency and reducing duplication.
 *
 * @example
 * ```typescript
 * import { mockDevice, mockSession } from '@tests/mocks';
 *
 * const wrapper = mountComponent(MyComponent, {
 *   piniaOptions: {
 *     initialState: {
 *       devices: { device: mockDevice },
 *       sessions: { sessions: [mockSession] }
 *     }
 *   }
 * });
 * ```
 */

export * from "./device";
export * from "./session";
export * from "./tag";
export * from "./publicKey";
export * from "./privateKey";
export * from "./firewallRule";
export * from "./webEndpoint";
export * from "./namespace";
export * from "./user";
export * from "./invitation";
