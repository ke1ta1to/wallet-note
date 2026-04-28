import { apiClient } from "@/api/client";
import { queryOptions } from "@tanstack/react-query";

export const categoriesQuery = (orgId: string) =>
  queryOptions({
    queryKey: ["orgs", orgId, "categories"],
    queryFn: async () => {
      const { data, error } = await apiClient.GET(
        "/orgs/{org_id}/categories",
        { params: { path: { org_id: orgId } } },
      );
      if (error) throw error;
      return data;
    },
  });
