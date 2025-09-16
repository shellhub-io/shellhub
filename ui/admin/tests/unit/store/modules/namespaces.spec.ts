import { describe, expect, it } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useNamespacesStore from "@admin/store/modules/namespaces";

describe("Namespaces Pinia Store", () => {
  setActivePinia(createPinia());
  const namespacesStore = useNamespacesStore();

  it("returns default state", () => {
    expect(namespacesStore.namespaces).toEqual([]);
    expect(namespacesStore.namespaceCount).toEqual(0);
  });
});
