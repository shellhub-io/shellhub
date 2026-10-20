import { describe, it, expect, beforeEach } from "vitest";
import { renderHook, waitFor } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace } from "@/tests/factories";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { useNavSectionTitle } from "../navSections";

function serveNamespace(legacy: boolean) {
  server.use(
    http.get("*/api/namespaces/:tenant", () =>
      HttpResponse.json(
        mockNamespace({
          settings: {
            session_record: false,
            connection_announcement: "",
            legacy,
            require_verified_device_key: false,
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
  serveNamespace(true);
});

describe("useNavSectionTitle", () => {
  it("places Sessions under Resources in a legacy namespace", async () => {
    serveNamespace(true);
    expect(await sectionTitle("/sessions")).toBe("Resources");
  });

  it("places Sessions under SSH in a new namespace", async () => {
    serveNamespace(false);
    expect(await sectionTitle("/sessions")).toBe("SSH");
  });

  it("names the legacy key pages SSH too", async () => {
    serveNamespace(true);
    expect(await sectionTitle("/sshkeys/public-keys")).toBe("SSH");
  });

  it.each(["/dashboard", "/account"])(
    "gives %s no section title",
    async (route) => {
      expect(await sectionTitle(route)).toBeUndefined();
    },
  );
});
