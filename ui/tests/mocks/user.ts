import { IUser } from "@/interfaces/IUser";

/**
 * Mock user data for testing.
 * Provides a basic user object with all required fields.
 */
export const mockUser: IUser = {
  id: "user-123",
  username: "testuser",
  email: "test@example.com",
  name: "Test User",
  role: "owner",
  tenant: "fake-tenant-data",
  token: "fake-jwt-token",
  user: "testuser",
  recovery_email: "test-recovery@example.com",
};
