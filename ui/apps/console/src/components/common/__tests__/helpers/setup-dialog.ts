import { vi, afterAll } from "vitest";

// jsdom does not implement showModal/close. Stub them so they set/remove the
// `open` attribute, which is what React Testing Library uses to resolve the
// `dialog` role.
const _origShowModal = HTMLDialogElement.prototype.showModal;
const _origClose = HTMLDialogElement.prototype.close;

HTMLDialogElement.prototype.showModal = vi.fn(function (
  this: HTMLDialogElement,
) {
  this.setAttribute("open", "");
});
HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
  this.removeAttribute("open");
});

afterAll(() => {
  HTMLDialogElement.prototype.showModal = _origShowModal;
  HTMLDialogElement.prototype.close = _origClose;
});
