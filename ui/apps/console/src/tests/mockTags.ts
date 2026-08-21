import { vi } from "vitest";
import { getTags } from "@/client/sdk.gen";
import { mockSdkResponse } from "./sdk";
import { mockTag } from "./factories";

export function mockTags(names: string[]) {
  vi.mocked(getTags).mockResolvedValue(
    mockSdkResponse(
      names.map((name) => mockTag({ name })),
      { "X-Total-Count": String(names.length) },
    ),
  );
}
