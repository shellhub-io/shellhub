import { describe, it, expect, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useNavSectionTitle } from "../navSections";

function setAccessMode(mode: "legacy" | "identity") {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: false,
            connection_announcement: "",
            ssh_access_mode: mode,
            ssh_legacy_allowed: true,
          },
        }),
      ),
    ),
  );
}

async function sectionTitle(route: string) {
  const queryClient = new QueryClient();
  const { result } = renderHook(() => useNavSectionTitle(route), {
    wrapper: createTestWrapper({ queryClient }),
  });
  await waitFor(() => expect(queryClient.isFetching()).toBe(0));
  return result.current;
}

beforeEach(() => {
  seedAuthStore();
  setAccessMode("legacy");
});

describe("useNavSectionTitle", () => {
  it("places Sessions under Resources in legacy mode", async () => {
    expect(await sectionTitle("/sessions")).toBe("Resources");
  });

  it("places Sessions under SSH in identity mode", async () => {
    setAccessMode("identity");
    expect(await sectionTitle("/sessions")).toBe("SSH");
  });

  it("names the legacy key pages SSH too", async () => {
    expect(await sectionTitle("/sshkeys/public-keys")).toBe("SSH");
  });

  it.each(["/dashboard", "/profile"])(
    "gives %s no section title",
    async (route) => {
      expect(await sectionTitle(route)).toBeUndefined();
    },
  );
});
