import { type ReactNode } from "react";
import { renderHook, type RenderHookOptions } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

interface RenderHookWithRouterOptions
  extends Omit<RenderHookOptions<unknown>, "wrapper"> {
  initialEntries?: string[];
}

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
