import { Component } from "vue";
import { mount, MountingOptions, ComponentMountingOptions } from "@vue/test-utils";
import { vi } from "vitest";
import { createVuetify } from "vuetify";
import { createTestingPinia, TestingOptions } from "@pinia/testing";
import { SnackbarInjectionKey } from "@/plugins/snackbar";

/**
 * Mock snackbar that can be used in tests.
 * Always injected automatically in all mounted components.
 */
export const mockSnackbar = {
  showSuccess: vi.fn(),
  showError: vi.fn(),
  showInfo: vi.fn(),
  showWarning: vi.fn(),
};

/**
 * Extended mounting options specific to ShellHub tests.
 */
export interface ShellHubMountOptions<T> extends Omit<MountingOptions<T>, "global"> {
  /**
   * Options to pass to createTestingPinia.
   * If not provided, uses createTestingPinia defaults.
   */
  piniaOptions?: Partial<TestingOptions>;

  /**
   * Whether to use shallow mount instead of mount. Default: false
   */
  shallow?: boolean;

  /**
   * Custom global configuration to merge with defaults
   */
  global?: MountingOptions<T>["global"];
}

/**
 * Helper function to mount Vue components with common ShellHub test configurations.
 *
 * This function simplifies component testing by providing default configurations for:
 * - Vuetify (always included)
 * - Pinia store (with createTestingPinia)
 * - Mock snackbar (always injected, accessible via mockSnackbar export)
 *
 * @example
 * ```typescript
 * // Basic usage with defaults (Vuetify + Testing Pinia + Mock Snackbar)
 * const wrapper = mountComponent(MyComponent);
 *
 * // With custom router
 * const wrapper = mountComponent(MyComponent, {
 *   global: {
 *     plugins: [router]
 *   },
 * });
 *
 * // With props
 * const wrapper = mountComponent(MyComponent, {
 *   props: { title: 'Test' }
 * });
 *
 * // With custom Pinia options
 * const wrapper = mountComponent(MyComponent, {
 *   piniaOptions: { stubActions: false }
 * });
 *
 * // With stubs
 * const wrapper = mountComponent(MyComponent, {
 *   stubs: { 'v-file-upload': true }
 * });
 *
 * // Shallow mount
 * const wrapper = mountComponent(MyComponent, { shallow: true });
 *
 * // Testing snackbar calls
 * await wrapper.find('[data-test="submit"]').trigger('click');
 * expect(mockSnackbar.showSuccess).toHaveBeenCalledWith('Success!');
 * ```
 *
 * @param component - The Vue component to mount
 * @param options - Mounting options including ShellHub-specific options
 * @returns VueWrapper instance
 */
export const mountComponent = <T extends Component>(
  component: T,
  options: ShellHubMountOptions<T> = {},
) => {
  const {
    piniaOptions,
    shallow = false,
    global = {},
    ...restOptions
  } = options;

  // Build plugins array - Vuetify and Testing Pinia are always included
  const plugins = [
    createVuetify(),
    createTestingPinia(piniaOptions),
    ...(global.plugins || []),
  ];

  // Build provide object - mockSnackbar is always injected
  const provide = {
    ...(global.provide || {}),
    [SnackbarInjectionKey]: mockSnackbar,
  };

  // Build final global config
  const finalGlobal = {
    ...global,
    plugins,
    provide,
  };

  // Build final mounting options
  const finalOptions: MountingOptions<T> = {
    ...restOptions,
    global: finalGlobal,
    shallow,
  };

  return mount(component, finalOptions as ComponentMountingOptions<T>);
};
