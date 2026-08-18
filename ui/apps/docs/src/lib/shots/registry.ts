import type { ShotDeclaration } from "./manifest";

/**
 * The component is loaded through Vite's SSR module graph and the integration
 * through Node's. Each graph gets its own module instance, so a module-level
 * array would be created twice and the integration would read the copy the
 * component never wrote to: an empty manifest and a green build. Hanging the
 * store off `globalThis` under a shared key is what makes both graphs agree.
 */
const REGISTRY: unique symbol = Symbol.for("@shellhub/docs:shots");

type Host = typeof globalThis & { [REGISTRY]?: ShotDeclaration[] };

function store(): ShotDeclaration[] {
  const host = globalThis as Host;
  host[REGISTRY] ??= [];

  return host[REGISTRY];
}

export function recordShot(declaration: ShotDeclaration): void {
  store().push(declaration);
}

export function collectShots(): ShotDeclaration[] {
  return [...store()];
}

export function resetShots(): void {
  store().length = 0;
}
