import axios from "axios";

const handleErrorInTests = (error: unknown) => {
  if (!axios.isAxiosError(error)) {
    console.error(error);
    return;
  }

  const requestUrl = (error.response?.request as { responseURL?: string })?.responseURL || "unknown URL";
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
