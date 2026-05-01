import { createFileRoute, redirect } from "@tanstack/react-router";
import { format } from "date-fns";

export const Route = createFileRoute("/_authed/orgs/$orgId/")({
  beforeLoad: ({ params }) => {
    throw redirect({
      to: "/orgs/$orgId/transactions",
      params: { orgId: params.orgId },
      search: { month: format(new Date(), "yyyy-MM") },
    });
  },
});
