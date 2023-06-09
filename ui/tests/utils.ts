/* eslint-disable import/prefer-default-export */
/**
 * This function transforms the methods of a given component object by wrapping
 * the asynchronous methods in a new layer of promises. It allows you to track
 * the promises created during the transformation by returning an array of these promises.
 *
 * @param component Vue component
 * @returns Array of promises
 */
export const promisify = (component) => {
  const promises: Array<Promise<unknown>> = [];

  Object.entries(component.methods).forEach(([name, func]) => {
    if (func instanceof Function && func.constructor.name === "AsyncFunction") {
      Object.defineProperty(component.methods, name, {
        async value(...args) {
          const promise = new Promise<unknown>((resolve) => {
            func.apply(this, args).then(resolve);
          });
          promises.push(promise);
          return promise;
        },
      });
    }
  });

  return promises;
};
