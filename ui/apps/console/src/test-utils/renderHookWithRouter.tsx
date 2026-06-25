import { type ReactNode } from "react";
import { renderHook, type RenderHookOptions } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

interface RenderHookWithRouterOptions
  extends Omit<RenderHookOptions<unknown>, "wrapper"> {
  initialEntries?: string[];
}

/**
 * Wraps `renderHook` with a MemoryRouter (React Router v7) and a QueryClient
 * provider so hooks that call `useSearchParams` (or any other router/query
 * hook) don't throw outside a routing context.
 *
 * @param hook - The hook callback to render.
 * @param options - Optional. Pass `initialEntries` to seed the URL.
 *
 * @example
 * ```ts
 * const { result } = renderHookWithRouter(() => useSearchParams(), {
 *   initialEntries: ["/?page=2"],
 * });
 * const [searchParams] = result.current;
 * expect(searchParams.get("page")).toBe("2");
 * ```
 */
export function renderHookWithRouter<Result>(
  hook: () => Result,
  { initialEntries, ...rest }: RenderHookWithRouterOptions = {},
) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={initialEntries}>{children}</MemoryRouter>
      </QueryClientProvider>
    );
  }

  return renderHook(hook, { wrapper: Wrapper, ...rest });
}
