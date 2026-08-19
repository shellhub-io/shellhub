import { renderHook, type RenderHookOptions } from "@testing-library/react";
import { createTestWrapper } from "./wrapper";

interface RenderHookWithRouterOptions
  extends Omit<RenderHookOptions<unknown>, "wrapper"> {
  initialEntries?: string[];
}

export function renderHookWithRouter<Result>(
  hook: () => Result,
  { initialEntries, ...rest }: RenderHookWithRouterOptions = {},
) {
  return renderHook(hook, {
    wrapper: createTestWrapper({ initialEntries: initialEntries ?? ["/"] }),
    ...rest,
  });
}
