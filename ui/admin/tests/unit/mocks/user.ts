import { IAdminUser } from "@admin/interfaces/IUser";

export const mockUser: IAdminUser = {
  id: "user-123",
  username: "testuser",
  name: "testuser",
  email: "test@example.com",
  created_at: "2024-01-01T00:00:00Z",
  last_login: "2024-01-10T12:00:00Z",
  email_marketing: false,
  status: "confirmed",
  max_namespaces: 3,
  namespacesOwned: 1,
  admin: true,
  preferences: {
    auth_methods: ["local"],
  },
};

export const mockUsers: IAdminUser[] = [
  { ...mockUser, id: "user-1", username: "alice", email: "alice@example.com" },
  { ...mockUser, id: "user-2", username: "bob", email: "bob@example.com" },
  { ...mockUser, id: "user-3", username: "charlie", email: "charlie@example.com" },
];
