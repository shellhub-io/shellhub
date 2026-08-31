import type { ReactNode } from "react";
import { renderHook, type RenderHookOptions } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";

function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
      mutations: { retry: false },
    },
  });
}

/**
 * Builds the provider wrapper a component needs under test: a query client and a router. The
 * client is fresh per call unless one is passed, so cache does not leak between cases.
 */
export function createTestWrapper(opts?: {
  queryClient?: QueryClient;
  initialEntries?: string[];
}) {
  const qc = opts?.queryClient ?? createTestQueryClient();

  return function TestWrapper({ children }: { children: ReactNode }) {
    const content = (
      <QueryClientProvider client={qc}>{children}</QueryClientProvider>
    );
    if (opts?.initialEntries) {
      return (
        <MemoryRouter initialEntries={opts.initialEntries}>
          {content}
        </MemoryRouter>
      );
    }
    return content;
  };
}

/**
 * Renders a hook inside that wrapper, for one that uses React Query.
 */
export function renderHookWithClient<Result>(
  hook: () => Result,
  opts?: Omit<RenderHookOptions<unknown>, "wrapper">,
) {
  return renderHook(hook, { wrapper: createTestWrapper(), ...opts });
}
