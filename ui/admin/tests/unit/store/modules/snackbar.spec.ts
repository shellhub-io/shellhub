import { describe, expect, it, beforeEach } from "vitest";
import { setActivePinia, createPinia } from "pinia";
import useSnackbarStore from "@admin/store/modules/snackbar";

describe("Snackbar Pinia Store", () => {
  let snackbarStore: ReturnType<typeof useSnackbarStore>;

  const SnackbarMessageAndContentTypeSuccess = { typeMessage: "success", typeContent: "" };
  const SnackbarMessageAndContentType = { typeMessage: "default", typeContent: "" };
  const SnackbarMessageAndContentTypeCopy = { typeMessage: "", typeContent: "copy" };
  const SnackbarMessageAndContentTypeDefault = { typeMessage: "", typeContent: "" };
  const SnackbarMessageNoContent = { typeMessage: "no-content", typeContent: "" };
  const SnackbarMessageLicenseRequired = { typeMessage: "licenseRequired", typeContent: "" };

  beforeEach(() => {
    setActivePinia(createPinia());
    snackbarStore = useSnackbarStore();
  });

  it("Returns snackbar default variables", () => {
    expect(snackbarStore.getSnackbarSuccess).toBe(false);
    expect(snackbarStore.getSnackbarError).toBe(false);
    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageAndContentTypeDefault);
    expect(snackbarStore.getSnackbarCopy).toBe(false);
  });

  it("Sets snackbar success message", () => {
    snackbarStore.snackbarMessageAndContentType = SnackbarMessageAndContentTypeSuccess;
    snackbarStore.snackbarSuccess = true;

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageAndContentTypeSuccess);
    expect(snackbarStore.getSnackbarSuccess).toBe(true);
  });

  it("Sets snackbar success default", () => {
    snackbarStore.showSnackbarSuccessDefault();

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageAndContentType);
    expect(snackbarStore.getSnackbarSuccess).toBe(true);
  });

  it("Sets snackbar no content", () => {
    snackbarStore.showSnackbarNoContent();

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageNoContent);
    expect(snackbarStore.getSnackbarSuccess).toBe(true);
  });

  it("Unsets snackbar success", () => {
    snackbarStore.unsetShowStatusSnackbarSuccess();

    expect(snackbarStore.getSnackbarSuccess).toBe(false);
  });

  it("Sets snackbar error loading/action", () => {
    snackbarStore.showSnackbarErrorAction("");

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual({ typeMessage: "action", typeContent: "" });
    expect(snackbarStore.getSnackbarError).toBe(true);
  });

  it("Sets snackbar error license", () => {
    snackbarStore.showSnackbarErrorLicense("");

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageLicenseRequired);
    expect(snackbarStore.getSnackbarError).toBe(true);
  });

  it("Sets snackbar error default", () => {
    snackbarStore.showSnackbarErrorDefault();

    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageAndContentType);
    expect(snackbarStore.getSnackbarError).toBe(true);
  });

  it("Unsets snackbar error", () => {
    snackbarStore.unsetShowStatusSnackbarError();

    expect(snackbarStore.getSnackbarError).toBe(false);
  });

  it("Sets snackbar copy", () => {
    snackbarStore.showSnackbarCopy("copy");

    expect(snackbarStore.getSnackbarCopy).toBe(true);
    expect(snackbarStore.getSnackbarMessageAndContentType).toEqual(SnackbarMessageAndContentTypeCopy);
  });

  it("Unsets snackbar copy", () => {
    snackbarStore.unsetShowStatusSnackbarCopy();

    expect(snackbarStore.getSnackbarCopy).toBe(false);
  });
});
