import { mount } from "@vue/test-utils";
import { expect, it } from "vitest";
import { promisify } from "./utils";

const sleep = (ms: number) => new Promise((r) => { setTimeout(r, ms); });

const Component = {
  template: `
    <div>
      <button @click="nonAsyncFunction" data-test="nonAsyncFunction"/>
      <button @click="asyncFunction" data-test="asyncFunction"/>
      <button @click="asyncFunctionNested" data-test="asyncFunctionNested"/>
      <button @click="asyncFunctionWithParams(1, 2, 3)" data-test="asyncFunctionWithParams"/>
    </div>
  `,
  data() {
    return {
      nonAsyncFunctionCalled: false,
      asyncFunctionCalled: false,
      asyncFunctionWithReturnCalled: false,
      asyncFunctionNestedCalled: false,
      asyncFunctionParams: [],
    };
  },
  methods: {
    nonAsyncFunction() {
      this.nonAsyncFunctionCalled = true;
    },
    async asyncFunction() {
      await sleep(100);
      this.asyncFunctionCalled = true;
    },
    async asyncFunctionNested() {
      await sleep(100);
      this.asyncFunctionNestedCalled = await this.asyncFunctionWithReturn();
    },
    async asyncFunctionWithReturn() {
      await sleep(100);
      this.asyncFunctionWithReturnCalled = true;
      return true;
    },
    async asyncFunctionWithParams(...args) {
      await sleep(100);
      this.asyncFunctionParams = args;
    },
  },
};

it("nonAsyncFunction", async () => {
  const promises = promisify(Component);
  const wrapper = mount(Component);

  await wrapper.find("button[data-test='nonAsyncFunction']").trigger("click");

  expect(promises).toHaveLength(0);
  expect(wrapper.vm.nonAsyncFunctionCalled).toBe(true);
});

it("asyncFunction", async () => {
  const promises = promisify(Component);
  const wrapper = mount(Component);

  await wrapper.find("button[data-test='asyncFunction']").trigger("click");

  expect(promises).toHaveLength(1);

  await Promise.all(promises);

  expect(wrapper.vm.asyncFunctionCalled).toBe(true);
});

it("asyncFunctionNested", async () => {
  const promises = promisify(Component);
  const wrapper = mount(Component);

  await wrapper.find("button[data-test='asyncFunctionNested']").trigger("click");

  expect(promises).toHaveLength(1);

  await Promise.all(promises);

  expect(wrapper.vm.asyncFunctionWithReturnCalled).toBe(true);
  expect(wrapper.vm.asyncFunctionNestedCalled).toBe(true);
});

it("asyncFunctionWithParams", async () => {
  const promises = promisify(Component);
  const wrapper = mount(Component);

  await wrapper.find("button[data-test='asyncFunctionWithParams']").trigger("click");

  expect(promises).toHaveLength(1);

  await Promise.all(promises);

  expect(wrapper.vm.asyncFunctionParams).toEqual([1, 2, 3]);
});
