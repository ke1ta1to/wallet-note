import { fetchAuthSession } from "aws-amplify/auth";
import { apiClient } from "./client";

apiClient.use({
  async onRequest({ request }) {
    const session = await fetchAuthSession();
    const token = session.tokens?.accessToken?.toString();
    if (token) {
      request.headers.set("Authorization", `Bearer ${token}`);
    }
    return request;
  },
});
