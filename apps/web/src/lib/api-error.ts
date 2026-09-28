import axios from "axios";

export function getErrorMessage(error: unknown, fallback: string) {
  if (axios.isAxiosError(error)) {
    const message = error.response?.data?.error;
    if (typeof message === "string" && message.length > 0) {
      return message;
    }
  }
  return fallback;
}
