import { apiClient } from "@/api/client";
import { queryOptions } from "@tanstack/react-query";

export const meOrgsQuery = queryOptions({
  queryKey: ["me", "orgs"],
  queryFn: async () => {
    const { data, error } = await apiClient.GET("/me/orgs");
    if (error) throw error;
    return data;
  },
});

export const getOrgQuery = (orgId: string) =>
  queryOptions({
    queryKey: ["orgs", orgId],
    queryFn: async () => {
      const { data, error } = await apiClient.GET("/orgs/{org_id}", {
        params: { path: { org_id: orgId } },
      });
      if (error) throw error;
      return data;
    },
  });
