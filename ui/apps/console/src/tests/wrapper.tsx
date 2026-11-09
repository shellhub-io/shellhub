import type { ReactNode } from "react";
import { renderHook, type RenderHookOptions } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { MemoryRouter } from "react-router-dom";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";

/**
 * A query client for one test: it neither retries nor waits between attempts, so a failing
 * request settles at once.
 */
export function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false, retryDelay: 0 },
      mutations: { retry: false },
    },
  });
}

/**
 * Builds the provider wrapper a component needs under test: a query client, the clipboard
 * provider every `useCopy` caller needs, and a router when given `initialEntries`. The client is
 * fresh per call unless one is passed, so cache does not leak between cases. The clipboard
 * provider adds its warning dialog to the tree once a copy cannot reach the clipboard. A suite
 * that mocks `@/hooks/useCopy` must keep the module's real exports, since the provider reads its
 * context from there, and a test that checks `useCopy` throws outside a provider cannot use this
 * wrapper.
 */
export function createTestWrapper(opts?: {
  queryClient?: QueryClient;
  initialEntries?: string[];
}) {
  const qc = opts?.queryClient ?? createTestQueryClient();

  return function TestWrapper({ children }: { children: ReactNode }) {
    const content = (
      <QueryClientProvider client={qc}>
        <ClipboardProvider>{children}</ClipboardProvider>
      </QueryClientProvider>
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
 * Renders a hook inside that wrapper, for one that uses React Query or the clipboard.
 */
export function renderHookWithClient<Result>(
  hook: () => Result,
  opts?: Omit<RenderHookOptions<unknown>, "wrapper">,
) {
  return renderHook(hook, { wrapper: createTestWrapper(), ...opts });
}
