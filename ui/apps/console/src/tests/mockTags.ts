import { vi } from "vitest";
import { getTags } from "@/client/sdk.gen";
import { mockSdkResponse } from "./sdk";
import { mockTag } from "./factories";

/**
 * Makes the tags call resolve to tags with the given names, with the count in the header the
 * paginated hook reads.
 */
export function mockTags(names: string[]) {
  vi.mocked(getTags).mockResolvedValue(
    mockSdkResponse(
      names.map((name) => mockTag({ name })),
      { "X-Total-Count": String(names.length) },
    ),
  );
}
