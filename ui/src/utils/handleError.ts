import { AxiosError } from "axios";

const isAxiosError = (error: unknown): error is AxiosError => (
  typeof error === "object"
    && error !== null
    && (error as AxiosError).isAxiosError !== undefined
);

const handleErrorInTests = (error: unknown) => {
  if (!isAxiosError(error)) {
    console.error(error);
    return;
  }

  const requestUrl = error.response?.request.responseURL;
  console.log(`${error.message}\nRequest URL: ${requestUrl}`);
};

const handleError = (error: unknown) => {
  if (process.env.NODE_ENV !== "test") {
    console.error(">>>>>>>>>>>>>>>>>>>>>");
    console.error(error);
    console.error(">>>>>>>>>>>>>>>>>>>>>");
  } else handleErrorInTests(error);
};

export default handleError;
