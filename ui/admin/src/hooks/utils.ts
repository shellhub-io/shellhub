// eslint-disable-next-line import/prefer-default-export
export const isHTML = (str: string) => !(str || "")
  .replace(/<([^>]+?)([^>]*?)>(.*?)<\/\1>/gi, "")
  .replace(/(<([^>]+)>)/gi, "")
  .trim();
