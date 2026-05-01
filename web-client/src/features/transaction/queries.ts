import { apiClient } from "@/api/client";
import { queryOptions } from "@tanstack/react-query";

export const transactionsQuery = (orgId: string, month: string) =>
  queryOptions({
    queryKey: ["orgs", orgId, "transactions", { month }],
    queryFn: async () => {
      const { data, error } = await apiClient.GET(
        "/orgs/{org_id}/transactions",
        {
          params: {
            path: { org_id: orgId },
            query: { month },
          },
        },
      );
      if (error) throw error;
      return data;
    },
  });
