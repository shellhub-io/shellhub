class LocalStorageMock implements Storage {
  private readonly storage = new Map<string, string>();

  get length() {
    return this.storage.size;
  }

  clear() {
    this.storage.clear();
  }

  getItem(key: string) {
    return this.storage.get(key) ?? null;
  }

  key(index: number) {
    const keys = Array.from(this.storage.keys());
    return keys[index] ?? null;
  }

  removeItem(key: string) {
    this.storage.delete(key);
  }

  setItem(key: string, value: string) {
    this.storage.set(key, value);
  }
}

const localStorageStub: Storage = new LocalStorageMock();

Object.defineProperty(globalThis, "localStorage", {
  value: localStorageStub,
  configurable: true,
  writable: true,
});

if (typeof window !== "undefined") {
  Object.defineProperty(window, "localStorage", {
    value: localStorageStub,
    configurable: true,
    writable: true,
  });
}

Object.defineProperty(globalThis, "Storage", {
  value: LocalStorageMock,
  configurable: true,
  writable: true,
});
