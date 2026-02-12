import { ITag } from "@/interfaces/ITags";

/**
 * Mock tag data for testing.
 * Provides a basic tag object with all required fields.
 */
export const mockTag: ITag = {
  name: "tag-1",
  tenant_id: "fake-tenant-data",
  created_at: "2026-01-08T00:00:00.000Z",
  updated_at: "2026-01-08T00:00:00.000Z",
};

/**
 * Mock tags array for testing lists.
 * Provides multiple tags for list/table testing scenarios.
 */
export const mockTags: ITag[] = [
  mockTag,
  { ...mockTag, name: "tag-2" },
];
