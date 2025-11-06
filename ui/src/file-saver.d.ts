declare module "file-saver" {
  export function saveAs(
    blob: Blob,
    filename?: string,
    disableAutoBOM?: boolean,
  ): void;
}
