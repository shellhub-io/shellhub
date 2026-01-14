import { createRouter, createWebHistory } from "vue-router";
import { routes as appRoutes } from "@/router";
import { routes as adminRoutes } from "@admin/router";

/**
 * Creates a clean router instance for testing purposes.
 *
 * This utility creates a fresh Vue Router instance with web history mode,
 * useful for isolating router state between tests. By default, it uses the
 * application's routes, but custom routes can be provided for specific test scenarios.
 *
 * @param routes - Optional array of route configurations. Defaults to application routes from @/router
 * @returns A new Vue Router instance with the specified routes
 *
 * @example
 * // Basic usage with default app routes
 * const router = createCleanRouter();
 * const wrapper = mount(MyComponent, {
 *   global: {
 *     plugins: [router]
 *   }
 * });
 *
 * @example
 * // With custom routes for testing
 * const customRoutes = [
 *   { path: '/', component: Home },
 *   { path: '/about', component: About }
 * ];
 * const router = createCleanRouter(customRoutes);
 *
 * @example
 * // Testing route navigation
 * const router = createCleanRouter();
 * await router.push('/devices');
 * const wrapper = mount(DevicesView, {
 *   global: { plugins: [router] }
 * });
 * expect(router.currentRoute.value.path).toBe('/devices');
 */
const createCleanRouter = (routes = appRoutes) => createRouter({
  history: createWebHistory(),
  routes,
});

/**
 * Creates a clean admin router instance for testing purposes.
 *
 * This utility creates a fresh Vue Router instance with web history mode and
 * the /admin/ base path, without navigation guards. Useful for testing admin
 * views in isolation without auth/license checks.
 *
 * @returns A new Vue Router instance with admin routes
 *
 * @example
 * // Basic usage for admin view tests
 * const router = createCleanAdminRouter();
 * await router.push('/dashboard');
 * const wrapper = mountComponent(Dashboard, {
 *   global: { plugins: [router] }
 * });
 */
export const createCleanAdminRouter = (routes = adminRoutes) => createRouter({
  history: createWebHistory("/admin/"),
  routes,
});

export default createCleanRouter;
