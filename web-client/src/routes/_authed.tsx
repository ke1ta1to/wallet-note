import { isAuthenticated } from "@/auth/guards";
import { Outlet, createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/_authed")({
  beforeLoad: async ({ location }) => {
    const authenticated = await isAuthenticated();
    if (!authenticated) {
      throw redirect({
        to: "/sign-in",
        search: { redirect_to: location.href },
      });
    }
  },
  component: () => <Outlet />,
});
