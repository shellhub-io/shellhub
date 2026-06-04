import "@testing-library/jest-dom/vitest";
import { afterEach } from "vitest";
import { cleanup } from "@testing-library/react";

afterEach(cleanup);

HTMLDialogElement.prototype.showModal = function (this: HTMLDialogElement) {
  this.setAttribute("open", "");
};

HTMLDialogElement.prototype.close = function (this: HTMLDialogElement) {
  this.removeAttribute("open");
};
